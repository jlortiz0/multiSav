package streamy

// #cgo pkg-config: libavformat libavcodec libavutil libswscale
// #include <libavutil/error.h>
// #include "stream.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"image"
	"image/color"
	"strings"
	"unsafe"
)

type AvVideoReader struct {
	ptr     *C.LibavReader
	buf     []byte
	target  float32
	counter float32
}

func NewAvVideoReader(file string, fps float32) (*AvVideoReader, error) {
	fName := C.CString(file)
	defer C.free(unsafe.Pointer(fName))
	var out AvVideoReader
	var v *C.LibavReader
	code := C.libavreader_new(fName, &v)
	if code != 0 {
		return nil, errHelper(code)
	}
	out.ptr = v
	p := C.libavreader_dimensions(v)
	out.buf = make([]byte, p.x*p.y*4)
	if fps != 0 {
		out.target = float32(C.libavreader_fps(v)) / fps
	} else {
		out.target = 1
	}
	code = C.libavreader_next(v, (*C.uint8_t)(&out.buf[0]))
	if code != 0 {
		return nil, errHelper(code)
	}
	return &out, nil
}

func (v *AvVideoReader) GetDimensions() (int32, int32) {
	out := C.libavreader_dimensions(v.ptr)
	return int32(out.x), int32(out.y)
}

func (v *AvVideoReader) Destroy() error {
	C.libavreader_destroy(v.ptr)
	return nil
}

func (v *AvVideoReader) Read() ([]color.RGBA, error) {
	v.counter += v.target
	for v.counter >= 1 {
		code := C.libavreader_next(v.ptr, (*C.uint8_t)(&v.buf[0]))
		if code != 0 {
			return nil, errHelper(code)
		}
		v.counter--
	}
	return unsafe.Slice((*color.RGBA)(unsafe.Pointer(&v.buf[0])), len(v.buf)/4), nil
}

func (v *AvVideoReader) Read8() ([]byte, error) {
	v.counter += v.target
	for v.counter >= 1 {
		code := C.libavreader_next(v.ptr, (*C.uint8_t)(&v.buf[0]))
		if code != 0 {
			return nil, errHelper(code)
		}
		v.counter--
	}
	return v.buf, nil
}

func errHelper(code C.int) error {
	var errbuf [64]byte
	C.av_strerror(code, (*C.char)(unsafe.Pointer(&errbuf[0])), C.AV_ERROR_MAX_STRING_SIZE)
	text := string(errbuf[:])
	ind := strings.IndexByte(text, 0)
	if ind != -1 {
		text = text[:ind]
	}
	return errors.New(text)
}

func GetVideoFrame(file string) (*image.RGBA, error) {
	fName := C.CString(file)
	defer C.free(unsafe.Pointer(fName))
	var v *C.LibavReader
	code := C.libavreader_new(fName, &v)
	if code != 0 {
		return nil, errHelper(code)
	}
	p := C.libavreader_dimensions(v)
	img := image.NewRGBA(image.Rect(0, 0, int(p.x), int(p.y)))
	code = C.libavreader_next(v, (*C.uint8_t)(unsafe.Pointer(&img.Pix[0])))
	if code != 0 {
		return nil, errHelper(code)
	}
	return img, nil
}
