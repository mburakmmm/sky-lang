package ffi

/*
#cgo pkg-config: libffi
#cgo CFLAGS: -I/opt/homebrew/opt/libffi/include
#cgo LDFLAGS: -L/opt/homebrew/opt/libffi/lib -lffi -ldl

#include <stdlib.h>
#include <dlfcn.h>
#include <ffi/ffi.h>
#include <string.h>

// Helper functions for FFI calls
void* sky_dlopen(const char* path) {
    return dlopen(path, RTLD_NOW | RTLD_GLOBAL);
}

void* sky_dlsym(void* handle, const char* name) {
    return dlsym(handle, name);
}

int sky_dlclose(void* handle) {
    return dlclose(handle);
}

char* sky_dlerror() {
    return dlerror();
}

// FFI call wrapper
typedef struct {
    void* fn_ptr;
    ffi_cif cif;
    ffi_type** arg_types;
    void** arg_values;
    void* result;
    int nargs;
} sky_ffi_call_ctx;

sky_ffi_call_ctx* sky_ffi_prep(void* fn_ptr, int nargs, int return_type) {
    sky_ffi_call_ctx* ctx = (sky_ffi_call_ctx*)malloc(sizeof(sky_ffi_call_ctx));
    ctx->fn_ptr = fn_ptr;
    ctx->nargs = nargs;
    ctx->arg_types = (ffi_type**)malloc(sizeof(ffi_type*) * nargs);
    ctx->arg_values = (void**)malloc(sizeof(void*) * nargs);

    // Determine return type
    ffi_type* ret_type;
    switch(return_type) {
        case 0: ret_type = &ffi_type_void; break;
        case 1: ret_type = &ffi_type_sint64; break;
        case 2: ret_type = &ffi_type_double; break;
        case 3: ret_type = &ffi_type_pointer; break;
        default: ret_type = &ffi_type_void;
    }

    return ctx;
}

void sky_ffi_set_arg(sky_ffi_call_ctx* ctx, int index, int type, void* value) {
    // Determine argument type
    switch(type) {
        case 1: // int
            ctx->arg_types[index] = &ffi_type_sint64;
            break;
        case 2: // float
            ctx->arg_types[index] = &ffi_type_double;
            break;
        case 3: // pointer/string
            ctx->arg_types[index] = &ffi_type_pointer;
            break;
        default:
            ctx->arg_types[index] = &ffi_type_sint64;
    }
    ctx->arg_values[index] = value;
}

int sky_ffi_call(sky_ffi_call_ctx* ctx, int return_type, void* result_ptr) {
    ffi_type* ret_type;
    switch(return_type) {
        case 0: ret_type = &ffi_type_void; break;
        case 1: ret_type = &ffi_type_sint64; break;
        case 2: ret_type = &ffi_type_double; break;
        case 3: ret_type = &ffi_type_pointer; break;
        default: ret_type = &ffi_type_void;
    }

    ffi_status status = ffi_prep_cif(&ctx->cif, FFI_DEFAULT_ABI, ctx->nargs,
                                     ret_type, ctx->arg_types);
    if (status != FFI_OK) {
        return -1;
    }

    ffi_call(&ctx->cif, FFI_FN(ctx->fn_ptr), result_ptr, ctx->arg_values);
    return 0;
}

void sky_ffi_cleanup(sky_ffi_call_ctx* ctx) {
    free(ctx->arg_types);
    free(ctx->arg_values);
    free(ctx);
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// TypeKind FFI tip çeşitleri
type TypeKind int

const (
	Void TypeKind = iota
	Int
	Float
	Pointer
	String
)

// Type C tipi
type Type struct {
	Kind TypeKind
	Size int
}

var (
	// Temel FFI tipleri
	VoidType    = Type{Kind: Void, Size: 0}
	IntType     = Type{Kind: Int, Size: 8}
	FloatType   = Type{Kind: Float, Size: 8}
	PointerType = Type{Kind: Pointer, Size: 8}
	StringType  = Type{Kind: String, Size: 8}
)

// Library yüklenmiş bir C kütüphanesini temsil eder
type Library struct {
	path    string
	handle  unsafe.Pointer
	symbols map[string]unsafe.Pointer
	mu      sync.RWMutex
}

// Symbol C fonksiyon sembolü
type Symbol struct {
	name     string
	ptr      unsafe.Pointer
	lib      *Library
	retType  Type
	argTypes []Type
}

// Load bir C kütüphanesini yükler
func Load(path string) (*Library, error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	handle := C.sky_dlopen(cpath)
	if handle == nil {
		errMsg := C.GoString(C.sky_dlerror())
		return nil, fmt.Errorf("failed to load library %s: %s", path, errMsg)
	}

	lib := &Library{
		path:    path,
		handle:  handle,
		symbols: make(map[string]unsafe.Pointer),
	}

	// Cleanup on GC
	runtime.SetFinalizer(lib, func(l *Library) {
		l.Close()
	})

	return lib, nil
}

// Symbol kütüphaneden bir sembol alır
func (l *Library) Symbol(name string) (*Symbol, error) {
	l.mu.RLock()
	if ptr, ok := l.symbols[name]; ok {
		l.mu.RUnlock()
		return &Symbol{
			name: name,
			ptr:  ptr,
			lib:  l,
		}, nil
	}
	l.mu.RUnlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	ptr := C.sky_dlsym(l.handle, cname)
	if ptr == nil {
		errMsg := C.GoString(C.sky_dlerror())
		return nil, fmt.Errorf("symbol not found %s: %s", name, errMsg)
	}

	l.mu.Lock()
	l.symbols[name] = ptr
	l.mu.Unlock()

	return &Symbol{
		name: name,
		ptr:  ptr,
		lib:  l,
	}, nil
}

// SetSignature fonksiyonun imzasını ayarlar
func (s *Symbol) SetSignature(retType Type, argTypes ...Type) *Symbol {
	s.retType = retType
	s.argTypes = argTypes
	return s
}

// Call C fonksiyonunu çağırır
func (s *Symbol) Call(args ...interface{}) (interface{}, error) {
	if len(args) != len(s.argTypes) {
		return nil, fmt.Errorf("argument count mismatch: expected %d, got %d",
			len(s.argTypes), len(args))
	}

	// Prep FFI call
	ctx := C.sky_ffi_prep(s.ptr, C.int(len(args)), C.int(s.retType.Kind))
	if ctx == nil {
		return nil, fmt.Errorf("failed to prepare FFI call")
	}
	defer C.sky_ffi_cleanup(ctx)

	// Set arguments
	for i, arg := range args {
		if err := setFFIArg(ctx, i, s.argTypes[i], arg); err != nil {
			return nil, err
		}
	}

	// Call function
	var result interface{}
	switch s.retType.Kind {
	case Void:
		if C.sky_ffi_call(ctx, 0, nil) != 0 {
			return nil, fmt.Errorf("FFI call failed")
		}
		result = nil

	case Int:
		var ret C.int64_t
		if C.sky_ffi_call(ctx, 1, unsafe.Pointer(&ret)) != 0 {
			return nil, fmt.Errorf("FFI call failed")
		}
		result = int64(ret)

	case Float:
		var ret C.double
		if C.sky_ffi_call(ctx, 2, unsafe.Pointer(&ret)) != 0 {
			return nil, fmt.Errorf("FFI call failed")
		}
		result = float64(ret)

	case Pointer, String:
		var ret unsafe.Pointer
		if C.sky_ffi_call(ctx, 3, unsafe.Pointer(&ret)) != 0 {
			return nil, fmt.Errorf("FFI call failed")
		}
		if s.retType.Kind == String && ret != nil {
			result = C.GoString((*C.char)(ret))
		} else {
			result = ret
		}

	default:
		return nil, fmt.Errorf("unsupported return type: %d", s.retType.Kind)
	}

	return result, nil
}

// setFFIArg sets an FFI argument
func setFFIArg(ctx *C.sky_ffi_call_ctx, index int, typ Type, value interface{}) error {
	switch typ.Kind {
	case Int:
		var val int64
		switch v := value.(type) {
		case int:
			val = int64(v)
		case int64:
			val = v
		case int32:
			val = int64(v)
		default:
			return fmt.Errorf("invalid int argument at index %d", index)
		}
		cval := C.int64_t(val)
		C.sky_ffi_set_arg(ctx, C.int(index), 1, unsafe.Pointer(&cval))

	case Float:
		var val float64
		switch v := value.(type) {
		case float64:
			val = v
		case float32:
			val = float64(v)
		default:
			return fmt.Errorf("invalid float argument at index %d", index)
		}
		cval := C.double(val)
		C.sky_ffi_set_arg(ctx, C.int(index), 2, unsafe.Pointer(&cval))

	case String:
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid string argument at index %d", index)
		}
		cstr := C.CString(str)
		// Note: This leaks memory - proper impl should track and free
		C.sky_ffi_set_arg(ctx, C.int(index), 3, unsafe.Pointer(&cstr))

	case Pointer:
		ptr, ok := value.(unsafe.Pointer)
		if !ok {
			return fmt.Errorf("invalid pointer argument at index %d", index)
		}
		C.sky_ffi_set_arg(ctx, C.int(index), 3, ptr)

	default:
		return fmt.Errorf("unsupported argument type at index %d", index)
	}

	return nil
}

// Close kütüphaneyi kapatır
func (l *Library) Close() error {
	if l.handle != nil {
		if C.sky_dlclose(l.handle) != 0 {
			errMsg := C.GoString(C.sky_dlerror())
			return fmt.Errorf("failed to close library: %s", errMsg)
		}
		l.handle = nil
	}
	return nil
}

// GetGlobalSymbol global bir sembol alır (main programdan)
func GetGlobalSymbol(name string) (*Symbol, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// RTLD_DEFAULT kullanarak main programdan sembol ara
	ptr := C.dlsym(C.RTLD_DEFAULT, cname)
	if ptr == nil {
		errMsg := C.GoString(C.sky_dlerror())
		return nil, fmt.Errorf("global symbol not found %s: %s", name, errMsg)
	}

	return &Symbol{
		name: name,
		ptr:  ptr,
	}, nil
}

// Struct for callback functions
type Callback struct {
	fn       func([]interface{}) interface{}
	retType  Type
	argTypes []Type
	cif      C.ffi_cif
	closure  *C.ffi_closure
}

// NewCallback creates a callback that can be called from C
func NewCallback(fn func([]interface{}) interface{}, retType Type, argTypes ...Type) (*Callback, error) {
	// This is a simplified version - full implementation would need more work
	return &Callback{
		fn:       fn,
		retType:  retType,
		argTypes: argTypes,
	}, nil
}

// Ptr returns the C function pointer for this callback
func (c *Callback) Ptr() unsafe.Pointer {
	// This would return the closure pointer
	return unsafe.Pointer(c.closure)
}

// Helper functions

// Malloc allocates memory using C malloc
func Malloc(size int) unsafe.Pointer {
	return C.malloc(C.size_t(size))
}

// Free frees memory allocated by Malloc
func Free(ptr unsafe.Pointer) {
	C.free(ptr)
}

// MemCopy copies memory
func MemCopy(dst, src unsafe.Pointer, size int) {
	C.memcpy(dst, src, C.size_t(size))
}

// CString converts Go string to C string (caller must free)
func CString(s string) unsafe.Pointer {
	return unsafe.Pointer(C.CString(s))
}

// GoString converts C string to Go string
func GoString(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	return C.GoString((*C.char)(ptr))
}

// Platform-specific helpers

// GetProcAddress Windows equivalent (Unix uses dlsym)
func (l *Library) GetProcAddress(name string) (unsafe.Pointer, error) {
	symbol, err := l.Symbol(name)
	if err != nil {
		return nil, err
	}
	return symbol.ptr, nil
}

// Error returns last dynamic linking error
func Error() string {
	return C.GoString(C.sky_dlerror())
}

// LoadedLibraries global registry
var loadedLibraries = make(map[string]*Library)
var loadedLibrariesMu sync.RWMutex

// GetLoadedLibrary gets a previously loaded library
func GetLoadedLibrary(path string) (*Library, bool) {
	loadedLibrariesMu.RLock()
	defer loadedLibrariesMu.RUnlock()
	lib, ok := loadedLibraries[path]
	return lib, ok
}

// RegisterLibrary registers a loaded library
func RegisterLibrary(lib *Library) {
	loadedLibrariesMu.Lock()
	defer loadedLibrariesMu.Unlock()
	loadedLibraries[lib.path] = lib
}
