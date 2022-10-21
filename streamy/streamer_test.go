package streamy_test

import (
	"image"
	"image/png"
	"os"
	"testing"

	"jlortiz.org/multisav/streamy"
)

func TestSomething(T *testing.T) {
	rd, err := streamy.NewAvVideoReader("iwonb.webm")
	if err != nil {
		T.Fatal(err)
	}
	sx, sy := rd.GetDimensions()
	T.Logf("Dimensions: %dx%d", sx, sy)
	data, err := rd.Read8()
	if err != nil {
		T.Fatal(err)
	}
	// f, err := os.OpenFile("out.tga", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	// if err != nil {
	// 	T.Fatal(err)
	// }
	// f.Write([]byte{0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(sx), byte(sx >> 8), byte(sy), byte(sy >> 8), 32, 8})
	// _, err = f.Write(data)
	// if err != nil {
	// 	T.Fatal(err)
	// }
	// f.Close()
	img := &image.RGBA{Pix: data, Stride: sx * 4, Rect: image.Rectangle{Max: image.Pt(sx, sy)}}
	f, err := os.OpenFile("out.png", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		T.Fatal(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		T.Fatal(err)
	}
	f.Close()
	rd.Destroy()
}
