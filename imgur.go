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
	if len(payload.Data.Images) == 0 {
		return nil, []ImageEntry{&payload.Data.ImgurImageEntry}
	} else if len(payload.Data.Images) == 1 {
		return nil, []ImageEntry{&payload.Data.Images[0]}
	}
	return nil, []ImageEntry{&ImgurGalleryEntry{
		ID: payload.Data.ID, Title: payload.Data.Title, Description: payload.Data.Description,
		Link: payload.Data.Description, Images: payload.Data.Images,
	}}
}

func (*ImgurSite) ExtendListing(_ interface{}) []ImageEntry { return nil }

type ImgurGalleryEntry struct {
	ID          string
	Title       string
	Description string
	Link        string
	Images      []ImgurImageEntry
}

func (*ImgurGalleryEntry) GetType() ImageEntryType { return IETYPE_GALLERY }

func (*ImgurGalleryEntry) GetText() string { return "" }

func (img *ImgurGalleryEntry) GetGalleryInfo(lazy bool) []ImageEntry {
	data := make([]ImageEntry, len(img.Images))
	if lazy {
		return data
	}
	for i := range data {
		data[i] = &img.Images[i]
	}
	return data
}

func (img *ImgurGalleryEntry) GetName() string { return img.Title }

func (img *ImgurGalleryEntry) GetInfo() string { return img.Title + "\n" + img.Description }

func (*ImgurGalleryEntry) GetDimensions() (int, int) { return -1, -1 }

func (img *ImgurGalleryEntry) GetPostURL() string { return "https://imgur.com/a/" + img.ID }

func (img *ImgurGalleryEntry) GetURL() string { return img.Link }

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

func (*ImgurImageEntry) GetGalleryInfo(bool) []ImageEntry { return nil }

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

type HybridImgurRedditGalleryEntry struct {
	RedditImageEntry
	imgur ImgurGalleryEntry
	title string
}

func (img *HybridImgurRedditGalleryEntry) GetInfo() string {
	return img.RedditImageEntry.GetInfo() + "\n\n" + img.imgur.GetInfo()
}

func (img *HybridImgurRedditGalleryEntry) GetURL() string {
	return img.imgur.GetURL()
}

func (entry *HybridImgurRedditGalleryEntry) GetName() string {
	if entry.title != "" {
		return entry.title
	}
	return entry.Title
}

func (*HybridImgurRedditGalleryEntry) GetType() ImageEntryType { return IETYPE_GALLERY }

func (img *HybridImgurRedditGalleryEntry) GetGalleryInfo(lazy bool) []ImageEntry {
	data := make([]ImageEntry, len(img.imgur.Images))
	if lazy {
		return data
	}
	for i := range data {
		data[i] = &HybridImgurRedditImageEntry{img.RedditImageEntry, img.imgur.Images[i],
			fmt.Sprintf("%s (%d/%d)", img.RedditImageEntry.Title, i+1, len(data)),
		}
	}
	return data
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
				switch x := w.(type) {
				case *ImgurImageEntry:
					data = append(data, &HybridImgurRedditImageEntry{*v.(*RedditImageEntry), *x, title})
				case *ImgurGalleryEntry:
					data = append(data, &HybridImgurRedditGalleryEntry{*v.(*RedditImageEntry), *x, title})
				}
			}
		} else {
			data = append(data, v)
		}
	}
	return data
}

func (img *HybridImgurRedditSite) GetListing(kind int, args []interface{}, persistent interface{}) (ImageListing, []ImageEntry) {
	inter, list := img.RedditSite.GetListing(kind, args, persistent)
	return inter, img.imgurRedditHybridHelper(list)
}

func (img *HybridImgurRedditSite) ExtendListing(cont ImageListing) []ImageEntry {
	list := img.RedditSite.ExtendListing(cont)
	return img.imgurRedditHybridHelper(list)
}

func NewHybridImgurRedditProducer(site *HybridImgurRedditSite, kind int, args []interface{}, persistent interface{}) *RedditProducer {
	return &RedditProducer{NewBufferedImageProducer(site, kind, args, persistent), &site.RedditSite}
}
