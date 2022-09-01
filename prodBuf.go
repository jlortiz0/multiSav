package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pkg/browser"
)

const BIP_BUFBEFORE = 5
const BIP_BUFAFTER = 5

type BufferedImageProducer struct {
	items     []ImageEntry
	listing   ImageListing
	site      ImageSite
	selSender chan int
	selRecv   chan bool
	lazy      bool
	buffer    [][]byte
	extending *sync.Once
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

func NewBufferedImageProducer(site ImageSite, kind int, args []interface{}, persistent interface{}) *BufferedImageProducer {
	buf := new(BufferedImageProducer)
	buf.site = site
	buf.listing, buf.items = site.GetListing(kind, args, persistent)
	if buf.items == nil {
		panic(buf.listing.(*ErrorListing).err)
	}
	buf.lazy = len(buf.items) != 0
	buf.buffer = make([][]byte, BIP_BUFAFTER+BIP_BUFBEFORE+1)
	buf.selSender = make(chan int, 3)
	buf.selRecv = make(chan bool)
	buf.extending = new(sync.Once)
	go func() {
		var prevSel int
		for {
			sel, ok := <-buf.selSender
			if !ok {
				break
			}
			if sel < prevSel {
				x := prevSel - sel
				copy(buf.buffer[minint(x, len(buf.buffer)):], buf.buffer)
				for i := 0; i < minint(x, len(buf.buffer)); i++ {
					buf.buffer[i] = nil
				}
			} else {
				x := sel - prevSel
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
				case "jpeg":
					fallthrough
				case "bmp":
					resp, err := http.Get(url)
					if err == nil {
						data, err := io.ReadAll(resp.Body)
						if err == nil {
							buf.buffer[i] = data
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
	buf.buffer = nil
}

func (buf *BufferedImageProducer) GetTitle() string {
	return "rediSav - Online"
}

func (buf *BufferedImageProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	if key == rl.KeyV {
		browser.OpenURL(buf.items[sel].GetURL())
	} else if key == rl.KeyH {
		browser.OpenURL(buf.items[sel].GetPostURL())
	} else if key == rl.KeyX {
		name := buf.items[sel].GetURL()
		ind := strings.IndexByte(name, '?')
		if ind != -1 {
			name = name[:ind]
		}
		ind = strings.LastIndexByte(name, '/')
		if ind != -1 {
			name = name[ind+1:]
		}
		name = "Downloads" + string(os.PathSeparator) + name
		if _, err := os.Stat(name); err == nil {
			i := 0
			ind = strings.LastIndexByte(name, '.')
			ext := name[ind+1:]
			name = name[:ind]
			for err == nil {
				i++
				_, err = os.Stat(fmt.Sprintf("%s_%d.%s", name, i, ext))
			}
			name = fmt.Sprintf("%s_%d.%s", name, i, ext)
		}
		if buf.buffer[BIP_BUFBEFORE] == nil {
			resp, err := http.Get(buf.items[sel].GetURL())
			if err == nil {
				f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
				if err == nil {
					io.Copy(f, resp.Body)
				}
				f.Close()
			} else {
				return ARET_NOTHING
			}
		} else {
			f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
			if err == nil {
				data := buf.buffer[BIP_BUFBEFORE]
				for len(data) > 0 {
					data2 := data
					if len(data) > 4096 {
						data2 = data2[:4096]
					}
					n, err := f.Write(data2)
					if err != nil {
						break
					}
					data = data[n:]
				}
			}
			f.Close()
		}
		buf.remove(sel)
		return ARET_MOVEUP | ARET_REMOVE
	}
	return ARET_NOTHING
}

func (buf *BufferedImageProducer) remove(sel int) {
	copy(buf.buffer[BIP_BUFBEFORE:], buf.buffer[BIP_BUFBEFORE+1:])
	buf.buffer[BIP_BUFAFTER+BIP_BUFBEFORE] = nil
	copy(buf.items[sel:], buf.items[sel+1:])
	buf.items = buf.items[:len(buf.items)-1]
}

func (buf *BufferedImageProducer) Get(sel int, img **rl.Image, ffmpeg **ffmpegReader) string {
	if sel+BIP_BUFAFTER >= len(buf.items) {
		go buf.extending.Do(func() {
			extend := buf.site.ExtendListing(buf.listing)
			if len(extend) == 0 {
				buf.lazy = false
			} else {
				buf.items = append(buf.items, extend...)
			}
			buf.extending = new(sync.Once)
		})
	}
	if sel >= len(buf.items) && buf.lazy {
		// If the above function finishes after this check but before here, the Do will become unusable
		// So we need to replace it anyway
		// Due to GC this shouldn't be an issue, and because Do will not return until it is done there should not be a write conflict
		buf.extending.Do(func() {})
		buf.extending = new(sync.Once)
		// Some might require multiple loads
		if sel >= len(buf.items) && buf.lazy {
			return buf.Get(sel, img, ffmpeg)
		}
	}
	if sel >= len(buf.items) && !buf.lazy {
		text := "You went too far!"
		vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
		*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
		return "\\/errPress left please"
	}
	switch buf.items[sel].GetType() {
	case IETYPE_GALLERY:
		data := buf.items[sel].GetGalleryInfo()
		if sel+1 != len(buf.items) {
			buf.items = append(buf.items[:sel+len(data)], buf.items[sel+1:]...)
			for i, x := range data {
				buf.items[sel+i] = x
			}
			if len(data) < BIP_BUFAFTER+1 {
				copy(buf.buffer[BIP_BUFBEFORE+len(data):], buf.buffer[BIP_BUFBEFORE+1:])
				for i := 0; i < len(data); i++ {
					buf.buffer[BIP_BUFBEFORE+i] = nil
				}
			} else {
				for i := BIP_BUFBEFORE; i < BIP_BUFBEFORE+BIP_BUFAFTER+1; i++ {
					buf.buffer[i] = nil
				}
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
		data := buf.buffer[BIP_BUFBEFORE]
		if data != nil {
			*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
			if (*img).Height == 0 {
				text := "Failed to load image?"
				vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
				*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
				rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
				return "\\/err" + buf.items[sel].GetName()
			}
		} else {
			s := minint(int((*ffmpeg).w), int((*ffmpeg).h))
			*img = rl.GenImageChecked(int((*ffmpeg).w), int((*ffmpeg).h), s/16, s/16, rl.Magenta, rl.Black)
		}
	case "png":
		fallthrough
	case "jpg":
		fallthrough
	case "jpeg":
		fallthrough
	case "bmp":
		data := buf.buffer[BIP_BUFBEFORE]
		if data == nil {
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
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				text := err.Error()
				vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
				*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
				rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
				return "\\/err" + buf.items[sel].GetName()
			}
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

func (buf *BufferedImageProducer) GetInfo(sel int) string {
	return buf.items[sel].GetInfo()
}
