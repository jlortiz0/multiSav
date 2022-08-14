package main

import (
	"io"
	"net/http"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
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
const BIP_BUFAFTER = 10
const BIP_BUFLOAD = true

type BufferedImageProducer struct {
	items     []string
	listing   interface{}
	site      ImageSite
	selSender chan int
	lazy      bool
	buffer    []*rl.Image
}

func NewBufferedImageProducer(kind int, args []interface{}, site ImageSite) *BufferedImageProducer {
	buf := new(BufferedImageProducer)
	buf.site = site
	buf.listing, buf.items = site.GetListing(kind, args)
	buf.lazy = len(buf.items) != 0
	buf.selSender = make(chan int, 3)
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
	return "heck if I know"
}

func (buf *BufferedImageProducer) ActionHandler(int32, int, int) ActionRet {
	return 0
}

func (buf *BufferedImageProducer) Get(sel int, img **rl.Image, ffmpeg **ffmpegReader) string {
	ind := strings.LastIndexByte(buf.items[sel], '.')
	ext := strings.ToLower(buf.items[sel][ind:])
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
		resp, err := http.Get(buf.items[sel])
		if err != nil {
			text := err.Error()
			vec := rl.MeasureTextEx(font, text, TEXT_SIZE, 0)
			*img = rl.GenImageColor(int(vec.X)+16, int(vec.Y)+10, rl.RayWhite)
			rl.ImageDrawTextEx(*img, rl.Vector2{X: 8, Y: 5}, font, text, TEXT_SIZE, 0, rl.Black)
			return "\\/err" + buf.items[sel]
		}
		f, _ := os.OpenFile("out.dat", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		defer f.Close()
		data, err := io.ReadAll(io.TeeReader(resp.Body, f))
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
