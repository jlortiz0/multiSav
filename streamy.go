package main

import (
	"image/color"
	"sync"
	"time"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jlortiz0/multisav/streamy"
)

type VideoReader interface {
	Destroy() error
	Read() ([]color.RGBA, *rl.Image, error)
	GetDimensions() (int32, int32)
}

type StreamyWrapperClass struct {
	*streamy.AvVideoReader
	buf     []color.RGBA
	slpTime *time.Ticker
	stop    chan struct{}
	lock    *sync.Mutex
}

func NewStreamy(f string) (*StreamyWrapperClass, error) {
	s, err := streamy.NewAvVideoReader(f)
	if err != nil {
		return nil, err
	}
	x, y := s.GetDimensions()
	buf := make([]color.RGBA, x*y)
	fps := s.GetFPS()
	if fps < 4 {
		fps = 4
	}
	return &StreamyWrapperClass{s, buf, time.NewTicker(time.Second / time.Duration(fps)), make(chan struct{}, 1), new(sync.Mutex)}, nil
}

func (s *StreamyWrapperClass) Read() ([]color.RGBA, *rl.Image, error) {
	select {
	case <-s.stop:
		return nil, nil, nil
	case <-s.slpTime.C:
	}
	other := unsafe.Slice((*byte)(unsafe.Pointer(&s.buf[0])), len(s.buf)*4)
	if !s.lock.TryLock() {
		return nil, nil, nil
	}
	err := s.AvVideoReader.Read(other)
	s.lock.Unlock()
	if err != nil {
		return nil, nil, err
	}
	return s.buf, nil, nil
}

func (s *StreamyWrapperClass) Destroy() error {
	s.stop <- struct{}{}
	close(s.stop)
	s.slpTime.Stop()
	s.lock.Lock()
	return s.AvVideoReader.Destroy()
}
