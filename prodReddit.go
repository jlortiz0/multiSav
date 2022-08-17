package main

import (
	"io"
	"net/http"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pkg/browser"
	"jlortiz.org/redisav/redditapi"
)

type RedditSite struct {
	redditapi.Reddit
}

func NewRedditSite(clientId, clientSecret, user, pass string) *RedditSite {
	red := redditapi.NewReddit("", clientId, clientSecret)
	if user != "" {
		red.Login(user, pass)
	}
	return &RedditSite{*red}
}

func (red *RedditSite) Destroy() {
	red.Logout()
}

func (red *RedditSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			name: "New: subreddit",
			args: []ListingArgument{
				{
					name: "Subreddit",
				},
			},
		},
		{
			name: "New: all",
			args: nil,
		},
	}
}

func (red *RedditSite) GetListing(kind int, args []interface{}) (interface{}, []string) {
	var iter *redditapi.SubmissionIterator
	var err error
	switch kind {
	case 0:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListNew(0)
		}
	case 1:
		iter, err = red.ListNew(0)
	}
	if err != nil {
		return err, nil
	}
	data := make([]string, 0, iter.Buffered())
	for !iter.NextRequiresFetch() {
		x, err := iter.Next()
		if err == nil {
			data = append(data, x.URL)
		}
	}
	return iter, data
}

func (red *RedditSite) ExtendListing(cont interface{}) []string {
	iter, ok := cont.(*redditapi.SubmissionIterator)
	if !ok {
		return nil
	}
	x, err := iter.Next()
	if err != nil || x == nil {
		return nil
	}
	data := make([]string, 1, iter.Buffered()+1)
	data[0] = x.URL
	for !iter.NextRequiresFetch() {
		x, err = iter.Next()
		if err == nil {
			data = append(data, x.URL)
		}
	}
	return data
}

const BIP_BUFBEFORE = 5
const BIP_BUFAFTER = 5
const BIP_BUFLOAD = true

type BufferedImageProducer struct {
	items     []string
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

func NewBufferedImageProducer(kind int, args []interface{}, site ImageSite) *BufferedImageProducer {
	buf := new(BufferedImageProducer)
	buf.site = site
	buf.listing, buf.items = site.GetListing(kind, args)
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
				url := buf.items[sel+i-BIP_BUFBEFORE]
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
		browser.OpenURL(buf.items[sel])
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
	ind := strings.LastIndexByte(buf.items[sel], '.')
	ext := strings.ToLower(buf.items[sel][ind:])
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
		*ffmpeg = NewFfmpegReader(buf.items[sel])
		*img = rl.GenImageColor(int((*ffmpeg).w), int((*ffmpeg).h), rl.Blank)
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
		resp, err := http.Get(buf.items[sel])
		if err != nil {
			text := err.Error()
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel]
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
			return "\\/err" + buf.items[sel]
		}
		*img = rl.LoadImageFromMemory(ext, data, int32(len(data)))
		if (*img).Height == 0 {
			text := "Failed to load image?"
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel]
		}
	default:
		text := "Image format not supported"
		vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
		*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
		rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
		return "\\/err" + buf.items[sel]
	}
	return buf.items[sel]
}
