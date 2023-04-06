package main

import "time"
import "io"
import "sync"
import "unsafe"
import "image/color"
import rl "github.com/gen2brain/raylib-go/raylib"
import jxl "github.com/jlortiz0/go-jxl-decoder"

type JxlWrapper struct {
	*jxl.JxlDecoder
	slpTime *time.Ticker
	stop    chan struct{}
	lock    *sync.Mutex
	r       io.ReadSeeker
	fst     bool
}

func NewJxlWrapper(r io.ReadSeeker) *JxlWrapper {
	j := new(JxlWrapper)
	j.r = r
	j.JxlDecoder = jxl.NewJxlDecoder(r)
	info, err := j.Info()
	if err != nil {
		j.JxlDecoder.Destroy()
		return nil
	}
	if info.FrameDelay == 0 {
		info.FrameDelay = time.Hour
	}
	j.slpTime = time.NewTicker(info.FrameDelay)
	j.stop = make(chan struct{})
	j.lock = new(sync.Mutex)
	j.fst = true
	return j
}

func (j *JxlWrapper) GetDimensions() (int32, int32) {
	info, _ := j.Info()
	return int32(info.W), int32(info.H)
}

func (j *JxlWrapper) Read() ([]color.RGBA, *rl.Image, error) {
	if j.fst {
		j.fst = false
	} else {
		select {
		case <-j.stop:
			return nil, nil, nil
		case <-j.slpTime.C:
		}
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
