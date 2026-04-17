package numbat

/*
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/linux_amd64 -lnumbat_cgo -lm -ldl -lpthread
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linux_arm64 -lnumbat_cgo -lm -ldl -lpthread
#cgo linux,riscv64 LDFLAGS: -L${SRCDIR}/lib/linux_riscv64 -lnumbat_cgo -lm -ldl -lpthread

#cgo freebsd,arm64 LDFLAGS: -L${SRCDIR}/lib/freebsd_arm64 -lnumbat_cgo -lm -lpthread

#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/darwin_amd64 -lnumbat_cgo -lm -framework Security -framework CoreFoundation -framework Foundation
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/darwin_arm64 -lnumbat_cgo -lm -framework Security -framework CoreFoundation -framework Foundation

#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/lib/windows_amd64 -lnumbat_cgo -lm -lws2_32 -luserenv -lbcrypt

#include <stdlib.h>
#include <stdbool.h>

typedef struct NumbatWrapper NumbatWrapper;

typedef struct NumbatResult {
char* out;
char* err;
double value;
bool is_quantity;
char* unit;
} NumbatResult;

extern NumbatWrapper* numbat_init();
extern NumbatResult numbat_interpret(NumbatWrapper* wrapper, const char* code);
extern char* numbat_set_variable(NumbatWrapper* wrapper, const char* name, double value, const char* unit);
extern void numbat_free_result(NumbatResult res);
extern void numbat_free_string(char* s);
extern void numbat_free(NumbatWrapper* wrapper);
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

type Context struct {
	wrapper *C.NumbatWrapper
	mu      sync.Mutex // Ensures thread-safe access to the Numbat environment
}

// Result packages both the formatted string and the raw float64 value
type Result struct {
	StringOutput string
	Value        float64
	IsQuantity   bool
	Unit         string
}

// NewContext initializes a new Numbat context with the prelude loaded
func NewContext() *Context {
	return &Context{
		wrapper: C.numbat_init(),
	}
}

// Interpret evaluates the provided code string thread-safely
func (c *Context) Interpret(code string) (Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))

	res := C.numbat_interpret(c.wrapper, cCode)
	defer C.numbat_free_result(res)

	if res.err != nil {
		return Result{}, errors.New(C.GoString(res.err))
	}

	var out string
	if res.out != nil {
		out = C.GoString(res.out)
	}

	var unit string
	if res.unit != nil {
		unit = C.GoString(res.unit)
	}

	return Result{
		StringOutput: out,
		Value:        float64(res.value),
		IsQuantity:   bool(res.is_quantity),
		Unit:         unit,
	}, nil
}

// SetVariable defines a variable in the Numbat Context thread-safely.
// Unit can be an empty string if you are setting a scalar value.
func (c *Context) SetVariable(name string, value float64, unit string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cUnit *C.char
	if unit != "" {
		cUnit = C.CString(unit)
		defer C.free(unsafe.Pointer(cUnit))
	}

	errStr := C.numbat_set_variable(c.wrapper, cName, C.double(value), cUnit)
	if errStr != nil {
		defer C.numbat_free_string(errStr)
		return errors.New(C.GoString(errStr))
	}

	return nil
}

// Free cleans up the context memory on the Rust side
func (c *Context) Free() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.wrapper != nil {
		C.numbat_free(c.wrapper)
		c.wrapper = nil
	}
}
