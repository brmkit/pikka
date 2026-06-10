//go:build !windows

package objload

/*
#cgo linux LDFLAGS: -ldl

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdarg.h>
#include <dlfcn.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__)
#define INTERNAL_DEFAULT_LIBRARY ((void*) -2)
#else
#define INTERNAL_DEFAULT_LIBRARY NULL
#endif

// === Beacon Compatibility Types ===

typedef struct {
	char *original;
	char *buffer;
	int   length;
	int   size;
} datap;

typedef struct {
	char *original;
	char *buffer;
	int   length;
	int   size;
} formatp;

// === Beacon Output Buffer ===

static char* _beacon_output = NULL;
static int _beacon_output_size = 0;
static int _beacon_output_offset = 0;

// === Beacon Data API ===

static void BeaconDataParse(datap* parser, char* buffer, int size) {
	if (parser == NULL) return;
	parser->original = buffer;
	parser->buffer = buffer;
	parser->length = size - 4;
	parser->size = size - 4;
	parser->buffer += 4;
}

static int BeaconDataInt(datap* parser) {
	int32_t val = 0;
	if (parser->length < 4) return 0;
	memcpy(&val, parser->buffer, 4);
	parser->buffer += 4;
	parser->length -= 4;
	return (int)val;
}

static short BeaconDataShort(datap* parser) {
	int16_t val = 0;
	if (parser->length < 2) return 0;
	memcpy(&val, parser->buffer, 2);
	parser->buffer += 2;
	parser->length -= 2;
	return (short)val;
}

static int BeaconDataLength(datap* parser) {
	return parser->length;
}

static char* BeaconDataExtract(datap* parser, int* size) {
	uint32_t length = 0;
	char* outdata = NULL;
	if (parser->length < 4) return NULL;
	memcpy(&length, parser->buffer, 4);
	parser->buffer += 4;
	outdata = parser->buffer;
	if (outdata == NULL) return NULL;
	parser->length -= 4;
	parser->length -= length;
	parser->buffer += length;
	if (size != NULL && outdata != NULL) *size = length;
	return outdata;
}

// === Beacon Format API ===

static void BeaconFormatAlloc(formatp* format, int maxsz) {
	if (format == NULL) return;
	format->original = calloc(maxsz, 1);
	format->buffer = format->original;
	format->length = 0;
	format->size = maxsz;
}

static void BeaconFormatReset(formatp* format) {
	memset(format->original, 0, format->size);
	format->buffer = format->original;
	format->length = format->size;
}

static void BeaconFormatFree(formatp* format) {
	if (format == NULL) return;
	if (format->original) {
		free(format->original);
		format->original = NULL;
	}
	format->buffer = NULL;
	format->length = 0;
	format->size = 0;
}

static void BeaconFormatAppend(formatp* format, char* text, int len) {
	memcpy(format->buffer, text, len);
	format->buffer += len;
	format->length += len;
}

static void BeaconFormatPrintf(formatp* format, char* fmt, ...) {
	va_list args;
	int length = 0;
	va_start(args, fmt);
	length = vsnprintf(NULL, 0, fmt, args);
	va_end(args);
	if (format->length + length > format->size) return;
	va_start(args, fmt);
	(void)vsnprintf(format->buffer, length, fmt, args);
	va_end(args);
	format->length += length;
	format->buffer += length;
}

static char* BeaconFormatToString(formatp* format, int* size) {
	*size = format->length;
	return format->original;
}

static uint32_t _swap_endianness(uint32_t indata) {
	uint32_t testint = 0xaabbccdd;
	uint32_t outint = indata;
	if (((unsigned char*)&testint)[0] == 0xdd) {
		((unsigned char*)&outint)[0] = ((unsigned char*)&indata)[3];
		((unsigned char*)&outint)[1] = ((unsigned char*)&indata)[2];
		((unsigned char*)&outint)[2] = ((unsigned char*)&indata)[1];
		((unsigned char*)&outint)[3] = ((unsigned char*)&indata)[0];
	}
	return outint;
}

static void BeaconFormatInt(formatp* format, int value) {
	uint32_t indata = value;
	uint32_t outdata = 0;
	if (format->length + 4 > format->size) return;
	outdata = _swap_endianness(indata);
	memcpy(format->buffer, &outdata, 4);
	format->length += 4;
	format->buffer += 4;
}

// === Beacon Output API ===

static void BeaconPrintf(int type, char* fmt, ...) {
	int length = 0;
	char* tempptr = NULL;
	va_list args;

	va_start(args, fmt);
	length = vsnprintf(NULL, 0, fmt, args) + 1;
	va_end(args);

	tempptr = realloc(_beacon_output, _beacon_output_size + length + 1);
	if (tempptr == NULL) return;
	_beacon_output = tempptr;
	memset(_beacon_output + _beacon_output_offset, 0, length + 1);

	va_start(args, fmt);
	length = vsnprintf(_beacon_output + _beacon_output_offset, length, fmt, args);
	_beacon_output_size += length;
	_beacon_output_offset += length;
	va_end(args);
}

static void BeaconOutput(int type, char* data, int len) {
	char* tempptr = NULL;
	tempptr = realloc(_beacon_output, _beacon_output_size + len + 1);
	if (tempptr == NULL) return;
	_beacon_output = tempptr;
	memset(_beacon_output + _beacon_output_offset, 0, len + 1);
	memcpy(_beacon_output + _beacon_output_offset, data, len);
	_beacon_output_size += len;
	_beacon_output_offset += len;
}

static int BeaconIsAdmin(void) {
	return 0;
}

extern char** environ;

static char** getEnviron(void) {
	return environ;
}

static char* getOSName(void) {
#ifdef __APPLE__
	return "apple";
#elif __FreeBSD__
	return "freebsd";
#elif __OpenBSD__
	return "openbsd";
#elif __linux__
	return "lin";
#else
	return "unk";
#endif
}

// === Internal Function Table ===

typedef struct {
	const char* functionName;
	void* function;
} beacon_func_entry_t;

static beacon_func_entry_t _beacon_functions[] = {
	{"BeaconDataParse",      (void*)BeaconDataParse},
	{"BeaconDataInt",        (void*)BeaconDataInt},
	{"BeaconDataShort",      (void*)BeaconDataShort},
	{"BeaconDataLength",     (void*)BeaconDataLength},
	{"BeaconDataExtract",    (void*)BeaconDataExtract},
	{"BeaconFormatAlloc",    (void*)BeaconFormatAlloc},
	{"BeaconFormatReset",    (void*)BeaconFormatReset},
	{"BeaconFormatFree",     (void*)BeaconFormatFree},
	{"BeaconFormatAppend",   (void*)BeaconFormatAppend},
	{"BeaconFormatPrintf",   (void*)BeaconFormatPrintf},
	{"BeaconFormatToString", (void*)BeaconFormatToString},
	{"BeaconFormatInt",      (void*)BeaconFormatInt},
	{"BeaconPrintf",         (void*)BeaconPrintf},
	{"BeaconOutput",         (void*)BeaconOutput},
	{"BeaconIsAdmin",        (void*)BeaconIsAdmin},
	{"getEnviron",           (void*)getEnviron},
	{"getOSName",            (void*)getOSName},
};

#define BEACON_FUNC_COUNT 17

// === Symbol Resolution ===

static void* _resolve_symbol(const char* symbolName) {
	int i;
	for (i = 0; i < BEACON_FUNC_COUNT; i++) {
		if (strcmp(symbolName, _beacon_functions[i].functionName) == 0) {
			return _beacon_functions[i].function;
		}
	}
	return dlsym(INTERNAL_DEFAULT_LIBRARY, symbolName);
}

// === BOF Entry Point Calling ===

typedef int (*bof_entry_t)(unsigned char*, int);

static int _call_bof(void* ptr, unsigned char* args, int args_len) {
	return ((bof_entry_t)ptr)(args, args_len);
}

// === Output Management ===

static void _reset_beacon_output(void) {
	if (_beacon_output) {
		free(_beacon_output);
		_beacon_output = NULL;
	}
	_beacon_output_size = 0;
	_beacon_output_offset = 0;
}

static char* _get_beacon_output(int* outsize) {
	char* out = _beacon_output;
	*outsize = _beacon_output_size;
	_beacon_output = NULL;
	_beacon_output_size = 0;
	_beacon_output_offset = 0;
	return out;
}
*/
import "C"
import "unsafe"

func resolveSymbol(name string) unsafe.Pointer {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C._resolve_symbol(cname)
}

func callBOFEntry(ptr unsafe.Pointer, args []byte) int {
	if len(args) > 0 {
		return int(C._call_bof(ptr, (*C.uchar)(unsafe.Pointer(&args[0])), C.int(len(args))))
	}
	return int(C._call_bof(ptr, nil, 0))
}

func resetBeaconOutput() {
	C._reset_beacon_output()
}

func getBeaconOutput() string {
	var outsize C.int
	cstr := C._get_beacon_output(&outsize)
	if cstr == nil || outsize == 0 {
		return ""
	}
	out := C.GoStringN(cstr, outsize)
	C.free(unsafe.Pointer(cstr))
	return out
}
