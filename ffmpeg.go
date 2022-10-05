/*
Copyright (C) 2019-2022 jlortiz

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"errors"
	"image"
	"image/color"
	"io"
	"os/exec"
	"regexp"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type VideoReader interface {
	Destroy() error
	Read() ([]color.RGBA, *rl.Image, error)
	GetDimensions() (int32, int32)
}

// TODO: This sucks. Use libavformat if possible.
type ffmpegReader struct {
	*exec.Cmd
	h, w   int32
	buf    []byte
	reader io.ReadCloser
}

func NewFfmpegReader(path string) *ffmpegReader {
	ffmpeg := new(ffmpegReader)
	fps := strconv.FormatInt(FRAME_RATE, 10)
	ffmpeg.Cmd = exec.Command("ffmpeg", "-stream_loop", "-1", "-i", path, "-r", fps,
		"-pix_fmt", "rgb24", "-vcodec", "rawvideo", "-f", "image2pipe", "-loglevel", "warning", "pipe:1")
	f, err := ffmpeg.StdoutPipe()
	if err != nil {
		panic(err)
	}
	ffmpeg.reader = f
	ffmpeg.h, ffmpeg.w = ffprobeFile(path)
	ffmpeg.buf = make([]byte, ffmpeg.h*ffmpeg.w*3)
	ffmpeg.Start()
	return ffmpeg
}

func NewFfmpegReaderKnownSize(path string, x, y int32) *ffmpegReader {
	ffmpeg := new(ffmpegReader)
	fps := strconv.FormatInt(FRAME_RATE, 10)
	ffmpeg.Cmd = exec.Command("ffmpeg", "-stream_loop", "-1", "-i", path, "-r", fps,
		"-pix_fmt", "rgb24", "-vcodec", "rawvideo", "-f", "image2pipe", "-loglevel", "warning", "pipe:1")
	f, err := ffmpeg.StdoutPipe()
	if err != nil {
		panic(err)
	}
	ffmpeg.reader = f
	ffmpeg.h, ffmpeg.w = y, x
	ffmpeg.buf = make([]byte, ffmpeg.h*ffmpeg.w*3)
	ffmpeg.Start()
	return ffmpeg
}

func (ffmpeg *ffmpegReader) Destroy() error {
	ffmpeg.reader.Close()
	ffmpeg.buf = nil
	err := ffmpeg.Process.Kill()
	if err != nil {
		return err
	}
	return ffmpeg.Wait()
}

func (ffmpeg *ffmpegReader) Read() ([]color.RGBA, *rl.Image, error) {
	_, err := io.ReadFull(ffmpeg.reader, ffmpeg.buf)
	if err != nil {
		return nil, nil, err
	}
	data2 := make([]color.RGBA, len(ffmpeg.buf)/3)
	for i := range data2 {
		data2[i] = color.RGBA{R: ffmpeg.buf[i*3], G: ffmpeg.buf[i*3+1], B: ffmpeg.buf[i*3+2], A: 255}
	}
	return data2, nil, nil
}

func (ffmpeg *ffmpegReader) GetDimensions() (int32, int32) {
	return ffmpeg.w, ffmpeg.h
}

var ffprobeRegex *regexp.Regexp = regexp.MustCompile(`Video: [^,]+, [^,].+, (\d+)x(\d+)`)

func ffprobeFile(path string) (int32, int32) {
	cmd := exec.Command("ffprobe", "-hide_banner", path)
	f, err := cmd.StderrPipe()
	if err != nil {
		return 0, 0
	}
	cmd.Start()
	data, _ := io.ReadAll(f)
	cmd.Wait()
	out := ffprobeRegex.FindSubmatch(data)
	if len(out) != 3 {
		return 0, 0
	}
	h, err := strconv.ParseInt(string(out[2]), 10, 32)
	if err != nil {
		return 0, 0
	}
	w, err := strconv.ParseInt(string(out[1]), 10, 32)
	if err != nil {
		return 0, 0
	}
	return int32(h), int32(w)
}

func GetVideoFrame(path string) (*image.RGBA, error) {
	h, w := ffprobeFile(path)
	if h < 1 || w < 1 {
		return nil, errors.New("dimensions too small")
	}
	cmd := exec.Command("ffmpeg", "-i", path, "-frames", "1", "-pix_fmt", "rgb0",
		"-vcodec", "rawvideo", "-f", "image2pipe", "-loglevel", "warning", "pipe:1")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	img := &image.RGBA{Pix: data, Stride: int(w) * 4, Rect: image.Rect(0, 0, int(w), int(h))}
	return img, nil
}
