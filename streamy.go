package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
	"jlortiz.org/multisav/streamy"
)

type VideoReader interface {
	Destroy() error
	Read() ([]color.RGBA, *rl.Image, error)
	GetDimensions() (int32, int32)
}

type StreamyWrapperClass streamy.AvVideoReader

func NewStreamy(f string) (*StreamyWrapperClass, error) {
	s, err := streamy.NewAvVideoReader(f, FRAME_RATE)
	return (*StreamyWrapperClass)(s), err
}

func (s *StreamyWrapperClass) Read() ([]color.RGBA, *rl.Image, error) {
	d, err := (*streamy.AvVideoReader)(s).Read()
	if err != nil {
		d = nil
	}
	return d, nil, err
}

func (s *StreamyWrapperClass) Destroy() error { return (*streamy.AvVideoReader)(s).Destroy() }

func (s *StreamyWrapperClass) GetDimensions() (int32, int32) {
	return (*streamy.AvVideoReader)(s).GetDimensions()
}
