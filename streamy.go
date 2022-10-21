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

type StreamyWrapperClass struct {
	streamy.AvVideoReader
}

func NewStreamy(f string) (StreamyWrapperClass, error) {
	s, err := streamy.NewAvVideoReader(f)
	return StreamyWrapperClass{s}, err
}

func (s StreamyWrapperClass) Read() ([]color.RGBA, *rl.Image, error) {
	d, err := s.AvVideoReader.Read()
	if err != nil {
		d = nil
	}
	return d, nil, err
}

func (s StreamyWrapperClass) GetDimensions() (int32, int32) {
	x, y := s.AvVideoReader.GetDimensions()
	return int32(x), int32(y)
}
