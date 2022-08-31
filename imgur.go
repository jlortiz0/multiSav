package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ImgurSite struct {
	key string
}

func NewImgurSite(key string) *ImgurSite {
	if !strings.HasPrefix(key, "Client-ID ") {
		key = "Client-ID " + key
	}
	return &ImgurSite{key: key}
}

func (*ImgurSite) Destroy() {}

func (*ImgurSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			name: "Album",
			args: []ListingArgument{
				{
					name: "ID",
				},
			},
		},
		{
			name: "Image",
			args: []ListingArgument{
				{
					name: "ID",
				},
			},
		},
	}
}

func (img *ImgurSite) GetListing(kind int, args []interface{}) (interface{}, []ImageEntry) {
	if kind == 2 {
		ind := strings.LastIndexByte(args[0].(string), '/')
		if args[0].(string)[ind-1] == 'a' {
			return img.GetListing(0, []interface{}{args[0].(string)[ind+1:]})
		}
		return img.GetListing(1, []interface{}{args[0].(string)[ind+1:]})
	}
	var url string
	if kind == 0 {
		url = "https://api.imgur.com/3/album/" + args[0].(string)
	} else {
		url = "https://api.imgur.com/3/imgur/" + args[0].(string)
	}
	rq, err := http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		return err, nil
	}
	rq.Header.Add("Authorization", img.key)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return err, nil
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(string(data)), nil
	}
	var payload struct {
		Data struct {
			ImgurImageEntry
			Images []ImgurImageEntry
		}
	}
	json.Unmarshal(data, &payload)
	if len(payload.Data.Images) != 0 {
		out := make([]ImageEntry, len(payload.Data.Images))
		for i := range payload.Data.Images {
			out[i] = &payload.Data.Images[i]
		}
		return nil, out
	}
	return nil, []ImageEntry{&payload.Data.ImgurImageEntry}
}

func (*ImgurSite) ExtendListing(_ interface{}) []ImageEntry { return nil }

type ImgurImageEntry struct {
	ID            string
	Title         string
	Description   string
	Link          string
	Mp4           string
	Width, Height int
}

func (*ImgurImageEntry) GetType() ImageEntryType { return IETYPE_REGULAR }

func (*ImgurImageEntry) GetText() string { return "" }

func (*ImgurImageEntry) GetGalleryInfo() []ImageEntry { return nil }

func (img *ImgurImageEntry) GetName() string { return img.Title }

func (img *ImgurImageEntry) GetInfo() string { return img.Title + "\n" + img.Description }

func (img *ImgurImageEntry) GetDimensions() (int, int) { return img.Width, img.Height }

func (img *ImgurImageEntry) GetPostURL() string { return "https://imgur.com/" + img.ID }

func (img *ImgurImageEntry) GetURL() string {
	if img.Mp4 != "" {
		return img.Mp4
	}
	return img.Link
}

type HybridImgurRedditImageEntry struct {
	RedditImageEntry
	imgur ImgurImageEntry
	title string
}

func (img *HybridImgurRedditImageEntry) GetInfo() string {
	return img.RedditImageEntry.GetInfo() + "\n\n" + img.imgur.GetInfo()
}

func (img *HybridImgurRedditImageEntry) GetURL() string {
	return img.imgur.GetURL()
}

func (entry *HybridImgurRedditImageEntry) GetName() string {
	if entry.title != "" {
		return entry.title
	}
	return entry.Title
}

type HybridImgurRedditSite struct {
	RedditSite
	imgur ImgurSite
}

func (img *HybridImgurRedditSite) imgurRedditHybridHelper(list []ImageEntry) []ImageEntry {
	data := make([]ImageEntry, 0, 2*len(list))
	for _, v := range list {
		if strings.HasSuffix(v.(*RedditImageEntry).Domain, "imgur.com") {
			if strings.ContainsRune(v.(*RedditImageEntry).URL[len(v.(*RedditImageEntry).URL)-6:], '.') {
				data = append(data, v)
				continue
			}
			err, out := img.imgur.GetListing(2, []interface{}{v.(*RedditImageEntry).URL})
			if err != nil {
				continue
			}
			for i, w := range out {
				title := ""
				if len(out) > 1 {
					title = fmt.Sprintf("%s (%d/%d)", v.(*RedditImageEntry).Title, i+1, len(out))
				}
				data = append(data, &HybridImgurRedditImageEntry{*v.(*RedditImageEntry), *w.(*ImgurImageEntry), title})
			}
		} else {
			data = append(data, v)
		}
	}
	return data
}

func (img *HybridImgurRedditSite) GetListing(kind int, args []interface{}) (interface{}, []ImageEntry) {
	inter, list := img.RedditSite.GetListing(kind, args)
	return inter, img.imgurRedditHybridHelper(list)
}

func (img *HybridImgurRedditSite) ExtendListing(cont interface{}) []ImageEntry {
	list := img.RedditSite.ExtendListing(cont)
	return img.imgurRedditHybridHelper(list)
}

func NewHybridImgurRedditProducer(site *HybridImgurRedditSite, kind int, args []interface{}) *RedditProducer {
	return &RedditProducer{NewBufferedImageProducer(site, kind, args), &site.RedditSite, kind, args}
}
