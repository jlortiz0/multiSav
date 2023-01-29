package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ImgurResolver struct {
	key string
}

func NewImgurResolver(key string) ImgurResolver {
	if !strings.HasPrefix(key, "Client-ID ") {
		key = "Client-ID " + key
	}
	return ImgurResolver{key: key}
}

func (ImgurResolver) GetResolvableDomains() []string {
	return []string{"imgur.com", "i.imgur.com", "www.imgur.com", "imgur.io"}
}

func (img ImgurResolver) ResolveURL(link string) (string, ImageEntry) {
	u, err := url.Parse(link)
	if err != nil {
		return "", nil
	}
	switch u.Hostname() {
	case "www.imgur.com":
		fallthrough
	case "imgur.io":
		fallthrough
	case "imgur.com":
		ind := strings.LastIndexByte(link, '/')
		if link[ind-1] == 'a' || link[ind-1] == 'y' {
			link = "https://api.imgur.com/3/album/" + link[ind+1:]
		} else {
			link = "https://api.imgur.com/3/image/" + link[ind+1:]
		}
		rq, err := http.NewRequest("GET", link, http.NoBody)
		if err != nil {
			return "", nil
		}
		rq.Header.Add("Authorization", img.key)
		resp, err := http.DefaultClient.Do(rq)
		if err != nil {
			return "", nil
		}
		if resp.StatusCode != 200 {
			return "", nil
		}
		data, _ := io.ReadAll(resp.Body)
		var payload struct {
			Data struct {
				ImgurImageEntry
				Images []ImgurImageEntry
			}
		}
		json.Unmarshal(data, &payload)
		if len(payload.Data.Images) == 0 {
			return "", &payload.Data.ImgurImageEntry
		} else if len(payload.Data.Images) == 1 {
			return "", &payload.Data.Images[0]
		}
		for i := range payload.Data.Images {
			payload.Data.Images[i].index = i + 1
			payload.Data.Images[i].parent = payload.Data.ID
		}
		return "", &ImgurGalleryEntry{
			ID: payload.Data.ID, Title: payload.Data.Title, Description: payload.Data.Description,
			Link: payload.Data.Link, Images: payload.Data.Images,
		}
	case "i.imgur.com":
		if strings.HasSuffix(link, ".gifv") {
			suff := link[strings.LastIndexByte(link, '/')+1:]
			return img.ResolveURL("https://imgur.com/" + suff[:len(suff)-5])
		}
		if strings.Contains(link, "/a/") {
			suff := link[strings.LastIndexByte(link, '/')+1:]
			ind := strings.LastIndexByte(suff, '.')
			if ind == -1 {
				ind = len(suff)
			}
			return img.ResolveURL("https://imgur.com/a/" + suff[:ind])
		}
		return RESOLVE_FINAL, nil
	}
	return "", nil
}

func (img ImgurResolver) GetRequest(link string) (*http.Response, error) {
	rq, err := http.NewRequest("GET", link, http.NoBody)
	if err != nil {
		return nil, err
	}
	// For some reason, adding this causes it to fail
	// Yet the function above works perfectly?
	// rq.Header.Add("Authorization", img.key)
	// fmt.Println(rq.Header)
	return http.DefaultClient.Do(rq)
}

type ImgurGalleryEntry struct {
	ID          string
	Title       string
	Description string
	Link        string
	Images      []ImgurImageEntry
}

func (*ImgurGalleryEntry) GetType() ImageEntryType { return IETYPE_GALLERY }

func (*ImgurGalleryEntry) GetText() string { return "" }

func (*ImgurGalleryEntry) Combine(ImageEntry) {}

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

func (img *ImgurGalleryEntry) GetInfo() string {
	if img.Title == "" {
		return img.Description
	} else if img.Description == "" {
		return img.Title
	}
	return img.Title + "\n" + img.Description
}

func (img *ImgurGalleryEntry) GetPostURL() string { return "https://imgur.com/a/" + img.ID }

func (img *ImgurGalleryEntry) GetURL() string { return img.Link }

func (img *ImgurGalleryEntry) GetSaveName() string { return img.Images[0].GetSaveName() }

type ImgurImageEntry struct {
	ID          string
	Title       string
	Description string
	Link        string
	Mp4         string
	index       int
	parent      string
}

func (*ImgurImageEntry) GetType() ImageEntryType { return IETYPE_REGULAR }

func (*ImgurImageEntry) GetText() string { return "" }

func (*ImgurImageEntry) Combine(ImageEntry) {}

func (*ImgurImageEntry) GetGalleryInfo(bool) []ImageEntry { return nil }

func (img *ImgurImageEntry) GetName() string { return img.Title }

func (img *ImgurImageEntry) GetInfo() string {
	if img.Title == "" {
		return img.Description
	} else if img.Description == "" {
		return img.Title
	}
	return img.Title + "\n" + img.Description
}

func (img *ImgurImageEntry) GetPostURL() string { return "https://imgur.com/" + img.ID }

func (img *ImgurImageEntry) GetURL() string {
	if img.Mp4 != "" {
		return img.Mp4
	}
	return img.Link
}

func (img *ImgurImageEntry) GetSaveName() string {
	s := img.GetURL()
	if img.parent == "" {
		ind := strings.LastIndexByte(s, '/')
		s = s[ind+1:]
		return s
	}
	ind := strings.LastIndexByte(s, '.')
	ext := s[ind+1:]
	return fmt.Sprintf("%s_%d.%s", img.parent, img.index, ext)
}
