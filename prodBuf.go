package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pkg/browser"
)

const BIP_BUFBEFORE = 5
const BIP_BUFAFTER = 5

type BufferedImageProducer struct {
	items     []ImageEntry
	listing   interface{}
	site      ImageSite
	selSender chan int
	selRecv   chan bool
	lazy      bool
	buffer    []*rl.Image
	extending bool
}

func minint(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxint(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func NewBufferedImageProducer(site ImageSite, kind int, args []interface{}) *BufferedImageProducer {
	buf := new(BufferedImageProducer)
	buf.site = site
	buf.listing, buf.items = site.GetListing(kind, args)
	if buf.items == nil {
		panic(buf.listing.(error))
	}
	buf.lazy = len(buf.items) != 0
	buf.buffer = make([]*rl.Image, BIP_BUFAFTER+BIP_BUFBEFORE+1)
	buf.selSender = make(chan int, 3)
	buf.selRecv = make(chan bool)
	go func() {
		var prevSel int
		for {
			sel, ok := <-buf.selSender
			if !ok {
				break
			}
			if sel < prevSel {
				x := prevSel - sel
				for i := maxint(len(buf.buffer)-x, 0); i < len(buf.buffer); i++ {
					if buf.buffer[i] != nil {
						rl.UnloadImage(buf.buffer[i])
					}
				}
				copy(buf.buffer[minint(x, len(buf.buffer)):], buf.buffer)
				for i := 0; i < minint(x, len(buf.buffer)); i++ {
					buf.buffer[i] = nil
				}
			} else {
				x := sel - prevSel
				for i := 0; i < minint(x, len(buf.buffer)); i++ {
					if buf.buffer[i] != nil {
						rl.UnloadImage(buf.buffer[i])
					}
				}
				copy(buf.buffer, buf.buffer[minint(x, len(buf.buffer)):])
				for i := maxint(len(buf.buffer)-x, 0); i < len(buf.buffer); i++ {
					buf.buffer[i] = nil
				}
			}
			buf.selRecv <- true
			for i := range buf.buffer {
				if sel+i-BIP_BUFBEFORE < 0 || buf.buffer[i] != nil || sel+i-BIP_BUFBEFORE+1 >= len(buf.items) {
					continue
				}
				url := buf.items[sel+i-BIP_BUFBEFORE].GetURL()
				ind := strings.LastIndexByte(url, '.')
				if ind == -1 {
					continue
				}
				ext := strings.ToLower(url[ind:])
				switch ext[1:] {
				case "png":
					fallthrough
				case "jpg":
					fallthrough
				case "bmp":
					resp, err := http.Get(url)
					if err == nil {
						data, err := io.ReadAll(resp.Body)
						if err == nil {
							buf.buffer[i] = rl.LoadImageFromMemory(ext, data, int32(len(data)))
						}
					}
				}
			}
			prevSel = sel
		}
	}()
	return buf
}

func (buf *BufferedImageProducer) IsLazy() bool { return buf.lazy }

func (buf *BufferedImageProducer) Len() int { return len(buf.items) }

func (buf *BufferedImageProducer) BoundsCheck(i int) bool {
	if i < 0 {
		return false
	}
	return buf.lazy || i < len(buf.items)
}

func (buf *BufferedImageProducer) Destroy() {
	close(buf.selSender)
	close(buf.selRecv)
}

func (buf *BufferedImageProducer) GetTitle() string {
	return "heck if I know"
}

func (buf *BufferedImageProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	if key == rl.KeyV {
		browser.OpenURL(buf.items[sel].GetURL())
	} else if key == rl.KeyH {
		browser.OpenURL(buf.items[sel].GetPostURL())
	}
	return ARET_NOTHING
}

func (buf *BufferedImageProducer) Get(sel int, img **rl.Image, ffmpeg **ffmpegReader) string {
	if sel+BIP_BUFAFTER >= len(buf.items) && !buf.extending {
		buf.extending = true
		go func() {
			extend := buf.site.ExtendListing(buf.listing)
			if len(extend) == 0 {
				buf.lazy = false
			} else {
				buf.items = append(buf.items, extend...)
			}
			buf.extending = false
		}()
	}
	switch buf.items[sel].GetType() {
	case IETYPE_GALLERY:
		data := buf.items[sel].GetGalleryInfo()
		if sel+1 != len(buf.items) {
			buf.items = append(buf.items[:sel+len(data)-1], buf.items[sel:]...)
			for i, x := range data {
				buf.items[sel+i] = x
			}
			for i := maxint(BIP_BUFBEFORE, len(buf.buffer)-len(data)); i < len(buf.buffer); i++ {
				if buf.buffer[i] != nil {
					rl.UnloadImage(buf.buffer[i])
				}
			}
			copy(buf.buffer[BIP_BUFBEFORE+minint(len(data), BIP_BUFAFTER+1):], buf.buffer[BIP_BUFBEFORE:])
			for i := 0; i < minint(len(data), BIP_BUFAFTER+1); i++ {
				buf.buffer[i+BIP_BUFBEFORE] = nil
			}
		} else {
			buf.items = append(buf.items, data...)
		}
	case IETYPE_TEXT:
		text := buf.items[sel].GetText()
		vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
		*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
		return buf.items[sel].GetName()
	}
	url := buf.items[sel].GetURL()
	ind := strings.LastIndexByte(url, '.')
	if ind == -1 {
		fmt.Println(buf.items[sel].GetName(), buf.items[sel].GetType())
		panic(url)
	}
	ext := strings.ToLower(url[ind:])
	buf.selSender <- sel
	<-buf.selRecv
	switch ext[1:] {
	case "mp4":
		fallthrough
	case "webm":
		fallthrough
	case "gif":
		fallthrough
	case "gifv":
		fallthrough
	case "mov":
		x, y := buf.items[sel].GetDimensions()
		if x != 0 {
			*ffmpeg = NewFfmpegReaderKnownSize(url, int32(x), int32(y))
		} else {
			*ffmpeg = NewFfmpegReader(url)
		}
		tmp := buf.buffer[BIP_BUFBEFORE]
		if tmp != nil {
			*img = rl.ImageCopy(tmp)
		} else {
			s := minint(int((*ffmpeg).w), int((*ffmpeg).h))
			*img = rl.GenImageChecked(int((*ffmpeg).w), int((*ffmpeg).h), s/16, s/16, rl.Magenta, rl.Black)
		}
	case "png":
		fallthrough
	case "jpg":
		fallthrough
	case "bmp":
		tmp := buf.buffer[BIP_BUFBEFORE]
		if tmp != nil {
			*img = rl.ImageCopy(tmp)
			break
		}
		resp, err := http.Get(url)
		if err != nil {
			text := err.Error()
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel].GetName()
		}
		// f, _ := os.OpenFile("out.dat", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		// defer f.Close()
		// data, err := io.ReadAll(io.TeeReader(resp.Body, f))
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			text := err.Error()
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel].GetName()
		}
		*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
		if (*img).Height == 0 {
			text := "Failed to load image?"
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel].GetName()
		}
	default:
		text := "Image format not supported\n" + url
		vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
		*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
		return "\\/err" + buf.items[sel].GetName()
	}
	return buf.items[sel].GetName()
}
