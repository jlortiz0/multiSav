package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
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
	bufLock   *sync.Mutex
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
	if site != nil {
		buf.listing, buf.items = site.GetListing(kind, args, persistent)
		// I forgot that some sites can legitmately return nil
		if _, ok := buf.listing.(ErrorListing); ok {
			buf.lazy = false
			buf.items = []ImageEntry{
				&TextImageEntry{
					wordWrapper(buf.listing.(ErrorListing).Error()),
					"\\/errError for you",
				},
			}
			buf.listing = nil
			buf.selSender = make(chan int, 1)
			return buf
		}
		buf.lazy = len(buf.items) != 0
	}
	buf.buffer = make([][]byte, BIP_BUFAFTER+BIP_BUFBEFORE+1)
	buf.bufLock = new(sync.Mutex)
	buf.selSender = make(chan int, 2)
	buf.selRecv = make(chan bool, 1)
	buf.extending = new(sync.Once)
	go func() {
		var prevSel int
		for {
			sel, ok := <-buf.selSender
			if !ok {
				close(buf.selRecv)
				buf.buffer = nil
				break
			}
			if sel < prevSel {
				x := prevSel - sel
				buf.bufLock.Lock()
				copy(buf.buffer[minint(x, len(buf.buffer)):], buf.buffer)
				for i := 0; i < minint(x, len(buf.buffer)); i++ {
					buf.buffer[i] = nil
				}
				buf.bufLock.Unlock()
			} else {
				x := sel - prevSel
				buf.bufLock.Lock()
				copy(buf.buffer, buf.buffer[minint(x, len(buf.buffer)):])
				for i := maxint(len(buf.buffer)-x, 0); i < len(buf.buffer); i++ {
					buf.buffer[i] = nil
				}
				buf.bufLock.Unlock()
			}
			buf.selRecv <- true
			for i := range buf.buffer {
				if sel+i-BIP_BUFBEFORE < 0 || buf.buffer[i] != nil || sel+i-BIP_BUFBEFORE+1 >= len(buf.items) {
					continue
				}
				URL := buf.items[sel+i-BIP_BUFBEFORE].GetURL()
				if buf.items[sel+i-BIP_BUFBEFORE].GetType() == IETYPE_GALLERY {
					ret := buf.items[sel+i-BIP_BUFBEFORE].GetGalleryInfo(false)
					if len(ret) == 0 {
						continue
					}
					URL = ret[0].GetURL()
				} else if buf.items[sel+i-BIP_BUFBEFORE].GetType() != IETYPE_REGULAR {
					continue
				}
				ind := strings.LastIndexByte(URL, '.')
				if ind == -1 {
					continue
				}
				ext := strings.ToLower(URL[ind:])
				switch ext[1:] {
				case "png":
					fallthrough
				case "jpg":
					fallthrough
				case "jpeg":
					fallthrough
				case "bmp":
					obj, _ := url.Parse(URL)
					if obj == nil {
						continue
					}
					resolve := resolveMap[obj.Host]
					var resp *http.Response
					var err error
					if resolve == nil {
						resp, err = http.DefaultClient.Get(URL)
					} else {
						resp, err = resolve.GetRequest(URL)
					}
					if err == nil {
						data, err := io.ReadAll(resp.Body)
						if err == nil {
							buf.bufLock.Lock()
							if buf.buffer[i] == nil {
								buf.buffer[i] = data
							}
							buf.bufLock.Unlock()
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
}

func (buf *BufferedImageProducer) GetTitle() string {
	return "multiSav - Online"
}

func (buf *BufferedImageProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	if key == rl.KeyV {
		browser.OpenURL(buf.items[sel].GetURL())
	} else if key == rl.KeyH {
		browser.OpenURL(buf.items[sel].GetPostURL())
	} else if key == rl.KeyC {
		buf.remove(sel)
		return ARET_MOVEDOWN | ARET_REMOVE
	} else if key == rl.KeyX {
		name := buf.items[sel].GetSaveName()
		if name == "" {
			return ARET_NOTHING
		}
		name = path.Join(saveData.Downloads, name)
		if _, err := os.Stat(name); err == nil {
			i := 0
			ind := strings.LastIndexByte(name, '.')
			ext := "png"
			if ind != -1 {
				ext = name[ind+1:]
				name = name[:ind]
			}
			for err == nil {
				i++
				_, err = os.Stat(fmt.Sprintf("%s_%d.%s", name, i, ext))
			}
			name = fmt.Sprintf("%s_%d.%s", name, i, ext)
		}
		if buf.buffer[BIP_BUFBEFORE] == nil {
			resp, err := buf.site.GetRequest(buf.items[sel].GetURL())
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
	} else if key == rl.KeyEnter {
		if buf.items[sel].GetType() == IETYPE_GALLERY {
			if call == 0 {
				return ARET_FADEOUT | ARET_AGAIN
			}
			menu := NewGalleryMenu(buf.items[sel], buf.site)
			ret := stdEventLoop(menu)
			menu.Destroy()
			if ret == LOOP_QUIT {
				return ARET_QUIT
			}
			rl.SetWindowTitle(buf.GetTitle())
			if ret == LOOP_EXIT {
				buf.remove(sel)
				return ARET_MOVEDOWN | ARET_REMOVE | ARET_FADEIN
			}
			return ARET_FADEIN
		}
	}
	return ARET_NOTHING
}

func (buf *BufferedImageProducer) remove(sel int) {
	buf.bufLock.Lock()
	copy(buf.buffer[BIP_BUFBEFORE:], buf.buffer[BIP_BUFBEFORE+1:])
	buf.buffer[BIP_BUFAFTER+BIP_BUFBEFORE] = nil
	buf.bufLock.Unlock()
	copy(buf.items[sel:], buf.items[sel+1:])
	buf.items = buf.items[:len(buf.items)-1]
}

func (buf *BufferedImageProducer) Get(sel int, img **rl.Image, ffmpeg *VideoReader) string {
	if sel+BIP_BUFAFTER >= len(buf.items) && buf.extending != nil && buf.lazy {
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
		*img = drawMessage("You went too far!")
		return "\\/errPress left please"
	}
	// The buffer should be kept updated even if we won't be reading it this time around
	buf.selSender <- sel
	if buf.items[sel].GetType() == IETYPE_TEXT {
		*img = drawMessage(buf.items[sel].GetText())
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return buf.items[sel].GetName()
	}
	URL := buf.items[sel].GetURL()
	if URL == "" {
		s := buf.items[sel].GetText()
		if s == "" {
			s = "Missing URL\n" + buf.items[sel].GetPostURL()
		}
		*img = drawMessage(s)
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return buf.items[sel].GetName()
	}
	_, ok := buf.items[sel].(*WrapperImageEntry)
	if !ok {
	Outer:
		for {
			ind := strings.LastIndexByte(URL, '.')
			if ind == -1 {
				ind = strings.Index(URL, "format=") + 6
				if ind == 5 {
					ind = len(URL) - 2
				}
			}
			ext := strings.ToLower(URL[ind:])
			if len(ext) > 4 {
				ind = strings.IndexByte(ext, '&')
				if ind != -1 {
					ext = ext[:ind]
				}
			}
			switch ext[1:] {
			case "mp4":
				fallthrough
			case "webm":
				fallthrough
			case "gif":
				fallthrough
			case "mov":
				fallthrough
			case "png":
				fallthrough
			case "jpg":
				fallthrough
			case "jpeg":
				fallthrough
			case "bmp":
				break Outer
			}
			domain, _ := url.Parse(URL)
			res, ok := resolveMap[domain.Hostname()]
			if !ok {
				break
			}
			newURL, newIE := res.ResolveURL(URL)
			if newURL == RESOLVE_FINAL {
				buf.items[sel] = &WrapperImageEntry{buf.items[sel], URL}
				break
			}
			if newIE != nil {
				URL = newIE.GetURL()
				buf.items[sel].Combine(newIE)
				break
			} else if newURL == "" {
				break
			}
			URL = newURL
		}
	}
	if buf.items[sel].GetType() == IETYPE_GALLERY {
		ret := buf.items[sel].GetGalleryInfo(false)
		if len(ret) == 0 {
			*img = drawMessage("This gallery is empty.")
			return "\\/err" + buf.items[sel].GetName()
		}
		URL = ret[0].GetURL()
	} else if buf.items[sel].GetType() == IETYPE_TEXT {
		*img = drawMessage(buf.items[sel].GetText())
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return buf.items[sel].GetName()
	} else if buf.items[sel].GetType() == IETYPE_UGOIRA {
		// TODO: DANGER DANGER SPECIAL CASING
		useful := buf.items[sel].(*PixivImageEntry)
		metadata, err := useful.GetUgoiraMetadata()
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + buf.items[sel].GetName()
		}
		resp, err := buf.site.GetRequest(metadata.Zip_urls.Best())
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + buf.items[sel].GetName()
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + buf.items[sel].GetName()
		}
		d := bytes.NewReader(data)
		r, _ := zip.NewReader(d, int64(len(data)))
		*ffmpeg = &UgoiraReader{
			h: int32(useful.Height), w: int32(useful.Width),
			frames: metadata.Frames, i: -1, reader: r}
	}
	ind2 := strings.LastIndexByte(URL, '?')
	if ind2 == -1 {
		ind2 = len(URL)
	}
	ind := strings.Index(URL[ind2:], "format=")
	if ind == -1 {
		ind = strings.LastIndexByte(URL[:ind2], '.')
		if ind == -1 {
			fmt.Println(buf.items[sel].GetName(), buf.items[sel].GetType())
			panic(fmt.Sprint(URL[:ind2], URL, ind2))
		}
	} else {
		ind += 6 + ind2
		ind2 = strings.IndexByte(URL[ind:], '&')
		if ind2 == -1 {
			ind2 = len(URL[ind:])
		}
		ind2 += ind
	}
	ext := strings.ToLower(URL[ind:ind2])
	if ext[0] == '=' {
		ext = "." + ext[1:]
	}
	<-buf.selRecv
	// TODO: copy changes to thread loader
	// TODO: handle &amp; (why is that in a URL in the first place?)
	switch ext[1:] {
	case "mp4":
		fallthrough
	case "webm":
		fallthrough
	case "gif":
		fallthrough
	case "mov":
		var err error
		*ffmpeg, err = NewStreamy(URL)
		if err != nil {
			panic(err)
		}
		data := buf.buffer[BIP_BUFBEFORE]
		if data != nil {
			*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
			if (*img).Height == 0 {
				*img = drawMessage("Failed to load image?")
				return "\\/err" + buf.items[sel].GetName()
			}
		} else {
			w, h := (*ffmpeg).GetDimensions()
			*img = rl.GenImageColor(int(w), int(h), rl.Black)
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
			obj, _ := url.Parse(URL)
			resolve := resolveMap[obj.Host]
			var resp *http.Response
			var err error
			if resolve == nil {
				resp, err = http.DefaultClient.Get(URL)
			} else {
				resp, err = resolve.GetRequest(URL)
			}
			if err != nil {
				*img = drawMessage(wordWrapper(err.Error()))
				return "\\/err" + buf.items[sel].GetName()
			}
			if resp.StatusCode/100 > 3 {
				URL = URL[:ind2]
				if resolve == nil {
					resp, err = http.DefaultClient.Get(URL)
				} else {
					resp, err = resolve.GetRequest(URL)
				}
				if err != nil {
					*img = drawMessage(wordWrapper(err.Error()))
					return "\\/err" + buf.items[sel].GetName()
				}
				if resp.StatusCode/100 > 3 {
					// Should I draw the body too?
					*img = drawMessage(resp.Status)
					return "\\/err" + buf.items[sel].GetName()
				}
			}
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				*img = drawMessage(wordWrapper(err.Error()))
				return "\\/err" + buf.items[sel].GetName()
			}
		}
		*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
		if (*img).Height == 0 {
			fmt.Println(string(data))
			*img = drawMessage("Failed to load image?\n" + URL)
			return "\\/err" + buf.items[sel].GetName()
		}
	default:
		*img = drawMessage("Image format not supported\n" + URL)
		return "\\/err" + buf.items[sel].GetName()
	}
	if buf.items[sel].GetType() == IETYPE_GALLERY {
		if *ffmpeg != nil {
			(*ffmpeg).Destroy()
			*ffmpeg = nil
			// rl.UnloadImage(*img)
			// text := fmt.Sprintf("Press Enter for gallery viewer (%d images)", len(buf.items[sel].GetGalleryInfo(true)))
			// vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			// *img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			// rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			// return "\\/err" + buf.items[sel].GetName()
		}
		size := float32((**img).Height) / 32
		if size < float32(font.BaseSize) {
			// This works around a quirk with ImageTextEx where it will only ever upscale text, never downscale it
			// I simply upscale the image first, correctness be damned
			// I could have made it integer upscale but that could cause the text to end up very small depending on the image
			scaleFactor := float64(font.BaseSize) / float64(size)
			rl.ImageResize(*img, int32(float64((**img).Width)*scaleFactor), int32(float64((**img).Height)*scaleFactor))
			size *= float32(scaleFactor)
		}
		text := fmt.Sprintf("Press Enter for gallery viewer (%d images)", len(buf.items[sel].GetGalleryInfo(true)))
		vec := rl.MeasureTextEx(font, text, size, 0)
		rl.ImageDrawRectangle(*img, 0, (**img).Height-int32(vec.Y)-10, (**img).Width, int32(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: (float32((**img).Width) - vec.X) / 2, Y: float32((**img).Height) - vec.Y - 5}, font, text, size, 0, rl.Black)
		return buf.items[sel].GetName()
	}
	return buf.items[sel].GetName()
}

func (buf *BufferedImageProducer) GetInfo(sel int) string {
	return buf.items[sel].GetInfo()
}

func (buf *BufferedImageProducer) GetListing() ImageListing {
	return buf.listing
}

func NewGalleryMenu(img ImageEntry, site ImageSite) *ImageMenu {
	prod := NewBufferedImageProducer(nil, 0, nil, nil)
	prod.items = img.GetGalleryInfo(false)
	prod.extending = nil
	prod.lazy = false
	prod.site = site
	menu := NewImageMenu(prod)
	return menu
}
