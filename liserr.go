package lisgo

/*
#include <libinsane/error.h>
#include <lislib.h>
*/
import "C"
import "unsafe"

//CErrorProxy wraps errors in C code, the idea is to reuse pre-allocated buffer when calling C function proxies
type CErrorProxy struct {
	cErr C.struct_error_proxy //Go allocates memory for this struct, so gc can move it in memory during cycles. Never store a pointer to this struct in C-memory
}

//NewErrProxy allocates internal *char buffer
func NewErrProxy() *CErrorProxy {
	var e CErrorProxy
	e.cErr.buf = (*C.char)(C.calloc(1000, C.sizeof_char))
	return &e

}

//FreeBuffer releases allocated memory
func (e *CErrorProxy) FreeBuffer() {
	if e.cErr.buf != nil {
		C.free(unsafe.Pointer(e.cErr.buf))
	}
}

//GetProxy returns address of internal struct error_proxy
func (e *CErrorProxy) GetProxy() *C.struct_error_proxy {
	return &e.cErr
}

//ErrNum returns number of error which we've got from C-code
func (e *CErrorProxy) ErrNum() uint32 {
	return e.cErr.err
}

func (e *CErrorProxy) Error() string {
	if e.cErr.buf == nil {
		return ""
	}
	return C.GoString(e.cErr.buf)
}

//Clear prepares CError for future use
func (e *CErrorProxy) Clear() {
	*e.cErr.buf = 0
	e.cErr.err = LisOk
}

//Naive leaky buffer implementation adapted from Effective Go.
//Cannot use sync.Pool as CError contains unmanaged memory buffer
// which must be released explicitly
var cErrorsPool = make(chan *CErrorProxy, 10)

func getErrorProxy() *CErrorProxy {

	var e *CErrorProxy
	// Grab a buffer if available; allocate if not.
	select {
	case e = <-cErrorsPool:
		// Got one; nothing more to do.
		e.Clear()
	default:
		// None free, so allocate a new one.
		e = NewErrProxy()
	}
	return e
}

func releaseErrorProxy(e *CErrorProxy) {
	select {
	case cErrorsPool <- e:
		// e returned to pool; nothing more to do.
	default:
		//pool is full, release it
		e.FreeBuffer()
	}
}
