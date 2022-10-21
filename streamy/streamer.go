package streamy

// #cgo pkg-config: libavformat libavcodec libavutil libswscale
// #include "stream.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"image/color"
	"strconv"
	"unsafe"
)

type VideoReader interface {
	Destroy() error
	Read() ([]byte, error)
	GetDimensions() (int32, int32)
}

type AvVideoReader struct {
	ptr *C.LibavReader
	buf []byte
	sz  uint32
}

func NewAvVideoReader(file string) (AvVideoReader, error) {
	fName := C.CString(file)
	defer C.free(unsafe.Pointer(fName))
	var out AvVideoReader
	var v *C.LibavReader
	code := C.libavreader_new(fName, &v)
	if code != 0 {
		return out, errors.New(strconv.Itoa(int(code)))
	}
	out.ptr = v
	p := C.libavreader_dimensions(v)
	out.buf = make([]byte, p.x*p.y*4)
	out.sz = uint32(p.x * p.y)
	return out, nil
}

func (v AvVideoReader) GetDimensions() (int, int) {
	out := C.libavreader_dimensions(v.ptr)
	return int(out.x), int(out.y)
}

func (v AvVideoReader) Destroy() error {
	C.libavreader_destroy(v.ptr)
	return nil
}

func (v AvVideoReader) Read() ([]color.RGBA, error) {
	code := C.libavreader_next(v.ptr, (*C.uint8_t)(&v.buf[0]))
	if code != 0 {
		return nil, errors.New(strconv.Itoa(int(code)))
	}
	return unsafe.Slice((*color.RGBA)(unsafe.Pointer(&v.buf[0])), v.sz), nil
}

func (v AvVideoReader) Read8() ([]byte, error) {
	code := C.libavreader_next(v.ptr, (*C.uint8_t)(&v.buf[0]))
	if code != 0 {
		return nil, errors.New(strconv.Itoa(int(code)))
	}
	return v.buf, nil
}
