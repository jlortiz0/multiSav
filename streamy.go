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
}

func NewStreamy(f string) (*StreamyWrapperClass, error) {
	s, err := streamy.NewAvVideoReader(f)
	if err != nil {
		return nil, err
	}
	x, y := s.GetDimensions()
	buf := make([]color.RGBA, x*y)
	fps := s.GetFPS()
	slpTime := FRAME_RATE / fps * float32(time.Second)
	return &StreamyWrapperClass{s, buf, time.NewTicker(time.Duration(slpTime))}, nil
}

func (s *StreamyWrapperClass) Read() ([]color.RGBA, *rl.Image, error) {
	<-s.slpTime.C
	other := unsafe.Slice((*byte)(unsafe.Pointer(&s.buf[0])), len(s.buf)*4)
	err := s.AvVideoReader.Read(other)
	if err != nil {
		return nil, nil, err
	}
	return s.buf, nil, nil
}

func (s *StreamyWrapperClass) Destroy() error {
	s.slpTime.Stop()
	return s.AvVideoReader.Destroy()
}
