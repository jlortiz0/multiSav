package main

import (
	"image/color"
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
}

func NewStreamy(f string) (*StreamyWrapperClass, error) {
	s, err := streamy.NewAvVideoReader(f)
	if err != nil {
		return nil, err
	}
	x, y := s.GetDimensions()
	buf := make([]color.RGBA, x*y)
	fps := s.GetFPS()
	slpTime := fps / FRAME_RATE * float32(time.Second)
	return &StreamyWrapperClass{s, buf, time.NewTicker(time.Duration(slpTime)), make(chan struct{}, 1)}, nil
}

// TODO: 33 fps gifs move at a snail's pace, why is that?
func (s *StreamyWrapperClass) Read() ([]color.RGBA, *rl.Image, error) {
	select {
	case <-s.stop:
		return s.buf, nil, nil
	case <-s.slpTime.C:
	}
	other := unsafe.Slice((*byte)(unsafe.Pointer(&s.buf[0])), len(s.buf)*4)
	err := s.AvVideoReader.Read(other)
	if err != nil {
		return nil, nil, err
	}
	return s.buf, nil, nil
}

func (s *StreamyWrapperClass) Destroy() error {
	s.stop <- struct{}{}
	s.slpTime.Stop()
	return s.AvVideoReader.Destroy()
}
