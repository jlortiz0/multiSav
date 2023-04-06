package main

import (
	"image/color"
	"io"
	"sync"
	"time"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"

	jxl "github.com/jlortiz0/go-jxl-decoder"
)

type JxlWrapper struct {
	*jxl.JxlDecoder
	slpTime *time.Timer
	stop    chan struct{}
	lock    *sync.Mutex
	r       io.ReadSeeker
}

func NewJxlWrapper(r io.ReadSeeker) *JxlWrapper {
	j := new(JxlWrapper)
	j.r = r
	j.JxlDecoder = jxl.NewJxlDecoder(r)
	_, err := j.Info()
	if err != nil {
		j.JxlDecoder.Destroy()
		return nil
	}
	j.slpTime = time.NewTimer(0)
	j.stop = make(chan struct{})
	j.lock = new(sync.Mutex)
	return j
}

func (j *JxlWrapper) GetDimensions() (int32, int32) {
	info, _ := j.Info()
	return int32(info.W), int32(info.H)
}

func (j *JxlWrapper) Read() ([]color.RGBA, *rl.Image, error) {
	select {
	case <-j.stop:
		return nil, nil, nil
	case <-j.slpTime.C:
	}
	buf, err := j.JxlDecoder.Read()
	if err != nil {
		return nil, nil, err
	}
	if buf == nil {
		j.Rewind()
		j.r.Seek(0, 0)
		buf, err = j.JxlDecoder.Read()
		if err != nil {
			return nil, nil, err
		}
	}
	other := unsafe.Slice((*color.RGBA)(unsafe.Pointer(&buf[0])), len(buf)/4)
	t := j.FrameDuration()
	if t == 0 {
		t = time.Hour
	}
	j.slpTime.Reset(t)
	return other, nil, nil
}

func (j *JxlWrapper) Destroy() error {
	close(j.stop)
	j.slpTime.Stop()
	j.lock.Lock()
	j.JxlDecoder.Destroy()
	c, ok := j.r.(io.ReadCloser)
	if ok {
		c.Close()
	}
	return nil
}
