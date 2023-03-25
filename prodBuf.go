package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pkg/browser"
	"github.com/rainycape/unidecode"
	"gitlab.com/tslocum/preallocate"
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
	buffer    []BufferObject
	extending *sync.Once
}

type BufferObject struct {
	URL  string
	data []byte
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
	buf.buffer = make([]BufferObject, BIP_BUFAFTER+BIP_BUFBEFORE+1)
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
				break
			}
			if sel < prevSel {
				x := prevSel - sel
				buf.bufLock.Lock()
				copy(buf.buffer[minint(x, len(buf.buffer)):], buf.buffer)
				for i := 0; i < minint(x, len(buf.buffer)); i++ {
					buf.buffer[i] = BufferObject{}
				}
				buf.bufLock.Unlock()
			} else {
				x := sel - prevSel
				buf.bufLock.Lock()
				copy(buf.buffer, buf.buffer[minint(x, len(buf.buffer)):])
				for i := maxint(len(buf.buffer)-x, 0); i < len(buf.buffer); i++ {
					buf.buffer[i] = BufferObject{}
				}
				buf.bufLock.Unlock()
			}
			buf.selRecv <- true
			for i := range buf.buffer {
				if sel+i-BIP_BUFBEFORE < 0 || sel+i-BIP_BUFBEFORE+1 >= len(buf.items) {
					continue
				}
				if len(buf.selSender) > 1 {
					break
				}
				URL := buf.items[sel+i-BIP_BUFBEFORE].GetURL()
				if buf.buffer[i].URL == URL {
					continue
				}
				if buf.items[sel+i-BIP_BUFBEFORE].GetType() == IETYPE_GALLERY {
					ret := buf.items[sel+i-BIP_BUFBEFORE].GetGalleryInfo(true)
					if len(ret) == 0 {
						continue
					}
					URL = ret[0].GetURL()
				} else if buf.items[sel+i-BIP_BUFBEFORE].GetType() == IETYPE_TEXT {
					continue
				}
				ind2 := strings.LastIndexByte(URL, '?')
				if ind2 == -1 {
					ind2 = len(URL)
				}
				ind := strings.Index(URL[ind2:], "format=")
				if ind == -1 {
					ind = strings.LastIndexByte(URL[:ind2], '.')
					if ind == -1 {
						continue
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
				if ext == "=png8" {
					ext = ".png"
				} else if ext == "=pjpg" {
					ext = ".jpg"
				} else if ext[0] == '=' {
					ext = "." + ext[1:]
				}
				if getExtType(ext[1:]) == EXT_PICTURE {
					obj, _ := url.Parse(URL)
					if obj == nil {
						continue
					}
					resolve := resolveMap[obj.Host]
					if resolve == nil {
						_, hst, ok := strings.Cut(obj.Host, ".")
						if !ok || !strings.ContainsRune(hst, '.') {
							break
						}
						resolve = resolveMap["*."+hst]
					}
					var resp *http.Response
					var err error
					if resolve == nil {
						resp, err = http.DefaultClient.Get(URL)
					} else {
						resp, err = resolve.GetRequest(URL)
					}
					if err == nil && resp.StatusCode/100 > 3 {
						ind2 := strings.LastIndexByte(URL, '?')
						if ind2 != -1 {
							URL = URL[:ind2]
							if resolve == nil {
								resp, err = http.DefaultClient.Get(URL)
							} else {
								resp, err = resolve.GetRequest(URL)
							}
						}
					}
					if err == nil && resp.StatusCode/100 < 3 {
						data, err := io.ReadAll(resp.Body)
						if err == nil {
							buf.bufLock.Lock()
							if buf.buffer[i].URL != URL {
								buf.buffer[i].URL = URL
								buf.buffer[i].data = data
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
	switch key {
	case rl.KeyV:
		browser.OpenURL(buf.items[sel].GetURL())
	case rl.KeyH:
		browser.OpenURL(buf.items[sel].GetPostURL())
	case rl.KeyC:
		buf.remove(sel)
		return ARET_MOVEDOWN | ARET_REMOVE
	case rl.KeyX:
		name := buf.items[sel].GetSaveName()
		if name == "" {
			return ARET_NOTHING
		}
		name = filepath.Join(saveData.Downloads, name)
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
		if buf.buffer[BIP_BUFBEFORE].URL != buf.items[sel].GetURL() {
			resp, err := buf.site.GetRequest(strings.ReplaceAll(buf.items[sel].GetURL(), "&amp;", "&"))
			if err == nil && resp.StatusCode == 200 {
				f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
				if err == nil {
					preallocate.File(f, resp.ContentLength)
					_, err = io.Copy(f, resp.Body)
					if err != nil {
						f.Close()
						return ARET_NOTHING
					}
				} else {
					return ARET_NOTHING
				}
				f.Close()
			} else {
				return ARET_NOTHING
			}
		} else {
			f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
			if err == nil {
				buf.bufLock.Lock()
				data := buf.buffer[BIP_BUFBEFORE].data
				buf.bufLock.Unlock()
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
			} else {
				return ARET_NOTHING
			}
			f.Close()
		}
		buf.remove(sel)
		return ARET_MOVEUP | ARET_REMOVE
	case rl.KeyEnter:
		if buf.items[sel].GetType() == IETYPE_GALLERY {
			if call == 0 {
				return ARET_FADEOUT | ARET_AGAIN
			}
			menu := NewGalleryMenu(buf.items[sel], buf.site, buf.buffer[BIP_BUFBEFORE].data)
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
	buf.buffer[BIP_BUFAFTER+BIP_BUFBEFORE] = BufferObject{}
	copy(buf.items[sel:], buf.items[sel+1:])
	buf.items = buf.items[:len(buf.items)-1]
	buf.bufLock.Unlock()
}

func (buf *BufferedImageProducer) Get(sel int, img **rl.Image, ffmpeg *VideoReader) string {
	if sel+BIP_BUFAFTER >= len(buf.items) && buf.extending != nil && buf.lazy {
		go buf.extending.Do(func() {
			extend, err := buf.site.ExtendListing(buf.listing)
			if len(extend) == 0 {
				buf.lazy = false
			} else {
				buf.items = append(buf.items, extend...)
			}
			if err != nil {
				buf.items = append(buf.items, &TextImageEntry{err.Error(), "Error for you"})
			}
			buf.extending = new(sync.Once)
		})
	}
	if sel >= len(buf.items) && buf.lazy {
		// If the above function finishes after this check but before here, the Do will become unusable
		// So we need to replace it anyway
		// Due to GC this shouldn't be an issue, and because Do will not return until it is done there should not be a write conflict
		// The sleep is there so that the thread gets a chance to start, otherwise this will go into an infinite loop until the scheduler can get to it
		time.Sleep(time.Millisecond * 50)
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
	current := buf.items[sel]
	URL := strings.ReplaceAll(current.GetURL(), "&amp;", "&")
	if current.GetType() == IETYPE_TEXT {
		*img = drawMessage(wordWrapper(unidecode.Unidecode(current.GetText())))
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return current.GetName()
	}
	if URL == "" {
		s := current.GetText()
		if s == "" {
			s = "Missing URL\n" + current.GetPostURL()
		}
		*img = drawMessage(unidecode.Unidecode(s))
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return current.GetName()
	}
	_, ok := current.(*WrapperImageEntry)
	if !ok {
		call := true
	Outer:
		for {
			domain, _ := url.Parse(URL)
			res, ok := resolveMap[domain.Hostname()]
			if !ok {
				_, hst, ok := strings.Cut(domain.Hostname(), ".")
				if !ok || !strings.ContainsRune(hst, '.') {
					break
				}
				res, ok = resolveMap["*."+hst]
				if !ok {
					break
				}
			}
			newURL, newIE := res.ResolveURL(URL)
			if newURL == RESOLVE_FINAL {
				buf.items[sel] = &WrapperImageEntry{current, URL, call}
				current = buf.items[sel]
				// <-buf.selRecv
				// buf.bufLock.Lock()
				// buf.buffer[BIP_BUFBEFORE] = BufferObject{}
				// buf.bufLock.Unlock()
				// go func() { buf.selRecv <- true }()
				break
			}
			call = false
			if newIE != nil {
				URL = strings.ReplaceAll(newIE.GetURL(), "&amp;", "&")
				tmp, _ := url.Parse(URL)
				current.Combine(newIE)
				if domain.Hostname() == tmp.Hostname() {
					break
				}
				continue
			} else if newURL == "" {
				break
			}
			URL = strings.ReplaceAll(newURL, "&amp;", "&")
			ind2 := strings.IndexByte(URL, '?')
			if ind2 == -1 {
				ind2 = len(URL)
			}
			ind := strings.LastIndexByte(URL[:ind2], '.')
			if ind == -1 {
				ind = strings.Index(URL, "format=") + 6
				if ind == 5 {
					ind = len(URL) - 2
				}
			}
			ext := strings.ToLower(URL[ind:])
			ind = strings.IndexByte(ext, '&')
			if ind != -1 {
				ext = ext[:ind]
			}
			if getExtType(ext[1:]) != EXT_NONE {
				buf.items[sel] = &WrapperImageEntry{current, URL, false}
				break Outer
			}
		}
	}
	switch current.GetType() {
	case IETYPE_GALLERY:
		ret := current.GetGalleryInfo(true)
		if len(ret) == 0 {
			*img = drawMessage("This gallery is empty.")
			return "\\/err" + current.GetName()
		}
		URL = ret[0].GetURL()
	case IETYPE_TEXT:
		*img = drawMessage(unidecode.Unidecode(current.GetText()))
		// We still need to recieve to ensure the buffer is updated, but no need to do it synchronously
		go func() { <-buf.selRecv }()
		return current.GetName()
	case IETYPE_UGOIRA:
		// DANGER DANGER SPECIAL CASING
		useful, ok := current.(PixivImageEntry)
		if !ok {
			useful = current.(*WrapperImageEntry).ImageEntry.(PixivImageEntry)
		}
		metadata, err := useful.GetUgoiraMetadata()
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + current.GetName()
		}
		resp, err := buf.site.GetRequest(metadata.Zip_urls.Best())
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + current.GetName()
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			*img = drawMessage(err.Error())
			return "\\/err" + current.GetName()
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
			fmt.Println(current.GetName(), current.GetType(), URL, ind2)
			*img = drawMessage("Unable to parse URL extension\n" + URL[:ind2])
			return "\\/err" + current.GetName()
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
	if ext == "=png8" {
		ext = ".png"
	} else if ext == "=pjpg" {
		ext = ".jpg"
	} else if ext[0] == '=' {
		ext = "." + ext[1:]
	}
	<-buf.selRecv
	switch getExtType(ext[1:]) {
	case EXT_VIDEO:
		var err error
		*ffmpeg, err = NewStreamy(URL)
		if err != nil {
			*ffmpeg = nil
			*img = drawMessage(wordWrapper(err.Error()))
			return "\\/err" + current.GetName()
		}
		data := buf.buffer[BIP_BUFBEFORE]
		if data.URL == URL {
			*img = rl.LoadImageFromMemory(ext, data.data, int32(len(data.data)))
			if (*img).Height == 0 {
				*img = drawMessage("Failed to load image?")
				return "\\/err" + current.GetName()
			}
		} else {
			w, h := (*ffmpeg).GetDimensions()
			*img = rl.GenImageColor(int(w), int(h), rl.Black)
		}
	case EXT_PICTURE:
		obj := buf.buffer[BIP_BUFBEFORE]
		data := obj.data
		if obj.URL != URL {
			obj, _ := url.Parse(URL)
			resolve := resolveMap[obj.Host]
			var resp *http.Response
			var err error
			if resolve == nil {
				_, hst, ok := strings.Cut(obj.Host, ".")
				if !ok || !strings.ContainsRune(hst, '.') {
					break
				}
				resolve = resolveMap["*."+hst]
			}
			if resolve == nil {
				resp, err = http.DefaultClient.Get(URL)
			} else {
				resp, err = resolve.GetRequest(URL)
			}
			if err != nil {
				*img = drawMessage(wordWrapper(err.Error()))
				return "\\/err" + current.GetName()
			}
			if resp.StatusCode/100 > 3 {
				body, _ := io.ReadAll(resp.Body)
				*img = drawMessage(wordWrapper(resp.Status + "\n" + string(body)))
				return "\\/err" + current.GetName()
			}
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				*img = drawMessage(wordWrapper(err.Error()))
				return "\\/err" + current.GetName()
			}
			if buf.bufLock.TryLock() {
				buf.buffer[BIP_BUFBEFORE] = BufferObject{URL, data}
				buf.bufLock.Unlock()
			}
		}
		*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
		if (*img).Height == 0 {
			fmt.Println(string(data))
			*img = drawMessage("Failed to load image?\n" + URL)
			return "\\/err" + current.GetName()
		}
	case EXT_NONE:
		*img = drawMessage("Image format not supported\n" + URL)
		return "\\/err" + current.GetName()
	}
	if current.GetType() == IETYPE_GALLERY {
		if *ffmpeg != nil {
			(*ffmpeg).Destroy()
			*ffmpeg = nil
		}
		size := float32((*img).Height) / 32
		if size < float32(font.BaseSize) {
			// This works around a quirk with ImageTextEx where it will only ever upscale text, never downscale it
			// I simply upscale the image first, correctness be damned
			// I could have made it integer upscale but that could cause the text to end up very small depending on the image
			scaleFactor := float64(font.BaseSize) / float64(size)
			rl.ImageResize(*img, int32(float64((*img).Width)*scaleFactor), int32(float64((*img).Height)*scaleFactor))
			size *= float32(scaleFactor)
		}
		text := fmt.Sprintf("Press Enter for gallery viewer (%d images)", len(current.GetGalleryInfo(true)))
		vec := rl.MeasureTextEx(font, text, size, 0)
		rl.ImageDrawRectangle(*img, 0, (*img).Height-int32(vec.Y)-10, (*img).Width, int32(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: (float32((*img).Width) - vec.X) / 2, Y: float32((*img).Height) - vec.Y - 5}, font, text, size, 0, rl.Black)
		return current.GetName()
	}
	return current.GetName()
}

func (buf *BufferedImageProducer) GetInfo(sel int) string {
	return unidecode.Unidecode(buf.items[sel].GetInfo())
}

func (buf *BufferedImageProducer) GetListing() ImageListing {
	return buf.listing
}

func NewGalleryMenu(img ImageEntry, site ImageSite, data []byte) *ImageMenu {
	menu := NewImageMenu(func() <-chan ImageProducer {
		ch := make(chan ImageProducer)
		go func() {
			prod := NewBufferedImageProducer(nil, 0, nil, nil)
			prod.items = img.GetGalleryInfo(false)
			prod.extending = nil
			prod.lazy = false
			prod.site = site
			if data != nil {
				bo := BufferObject{prod.items[0].GetURL(), data}
				prod.bufLock.Lock()
				prod.buffer[BIP_BUFBEFORE] = bo
				prod.bufLock.Unlock()
			}
			ch <- prod
		}()
		return ch
	})
	return menu
}
