//go:build windows
// +build windows

package exit

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	FileRenameInfo        = 3
	FileDispositionInfo   = 4
	FileDispositionInfoEx = 21
)

type FILE_RENAME_INFO struct {
	ReplaceIfExists bool
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}

type FILE_DISPOSITION_INFO struct {
	DeleteFile bool
}

type FILE_DISPOSITION_INFO_EX struct {
	Flags uint32
}

const (
	FILE_DISPOSITION_FLAG_DELETE          = 0x00000001
	FILE_DISPOSITION_FLAG_POSIX_SEMANTICS = 0x00000002
)

func selfDelete() error {
	// method: https://github.com/LloydLabs/delete-self-poc
	// 1. Open the file with DELETE access

	path, err := os.Executable()
	if err != nil {
		return err
	}

	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	hCurrent, err := windows.CreateFile(
		pathPtr,
		windows.DELETE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(hCurrent)

	// 2. Rename the file stream to :pikka
	streamName := ":pikka"
	streamNameUTF16, err := windows.UTF16FromString(streamName) // Includes null terminator
	if err != nil {
		return err
	}

	// Calculate size for FILE_RENAME_INFO
	// We allocate enough space for the struct + the string
	bufferSize := unsafe.Sizeof(FILE_RENAME_INFO{}) + uintptr(len(streamNameUTF16)*2)
	buffer := make([]byte, bufferSize)

	renameInfo := (*FILE_RENAME_INFO)(unsafe.Pointer(&buffer[0]))
	renameInfo.ReplaceIfExists = false
	renameInfo.RootDirectory = 0
	// FileNameLength should be in bytes, excluding null terminator
	renameInfo.FileNameLength = uint32((len(streamNameUTF16) - 1) * 2)

	// Copy the string into the buffer at the FileName offset
	startPtr := uintptr(unsafe.Pointer(&renameInfo.FileName[0]))
	for i, v := range streamNameUTF16 {
		*(*uint16)(unsafe.Pointer(startPtr + uintptr(i*2))) = v
	}

	// Call SetFileInformationByHandle
	err = setFileInformationByHandle(hCurrent, FileRenameInfo, unsafe.Pointer(&buffer[0]), uint32(bufferSize))
	if err != nil {
		return err
	}

	// 3. Close the handle (defer will do it, but we need it closed before re-opening)
	windows.CloseHandle(hCurrent)

	// 4. Open the file again
	hCurrent, err = windows.CreateFile(
		pathPtr,
		windows.DELETE|windows.SYNCHRONIZE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(hCurrent)

	// 5. Set Disposition Info
	// Try FileDispositionInfoEx first (Windows 10 1709+)
	dispInfoEx := FILE_DISPOSITION_INFO_EX{
		Flags: FILE_DISPOSITION_FLAG_DELETE | FILE_DISPOSITION_FLAG_POSIX_SEMANTICS,
	}
	err = setFileInformationByHandle(hCurrent, FileDispositionInfoEx, unsafe.Pointer(&dispInfoEx), uint32(unsafe.Sizeof(dispInfoEx)))
	if err != nil {
		// Fallback to standard FileDispositionInfo
		dispInfo := FILE_DISPOSITION_INFO{
			DeleteFile: true,
		}
		err = setFileInformationByHandle(hCurrent, FileDispositionInfo, unsafe.Pointer(&dispInfo), uint32(unsafe.Sizeof(dispInfo)))
		if err != nil {
			return err
		}
	}

	// 6. Close handle to trigger deletion (defer will do it)
	return nil
}

func setFileInformationByHandle(handle windows.Handle, fileInformationClass uint32, buf unsafe.Pointer, bufSize uint32) error {
	modkernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procSetFileInformationByHandle := modkernel32.NewProc("SetFileInformationByHandle")
	r1, _, e1 := procSetFileInformationByHandle.Call(
		uintptr(handle),
		uintptr(fileInformationClass),
		uintptr(buf),
		uintptr(bufSize),
	)
	if r1 == 0 {
		return e1
	}
	return nil
}
