package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ImgurResolver struct {
	key string
}

func NewImgurResolver(key string) *ImgurResolver {
	if !strings.HasPrefix(key, "Client-ID ") {
		key = "Client-ID " + key
	}
	return &ImgurResolver{key: key}
}

func (*ImgurResolver) Destroy() {}

func (*ImgurResolver) GetResolvableDomains() []string {
	return []string{"imgur.com", "i.imgur.com"}
}

func (img *ImgurResolver) ResolveURL(URL string) (string, ImageEntry) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", nil
	}
	switch u.Hostname() {
	case "imgur.com":
		ind := strings.LastIndexByte(URL, '/')
		if URL[ind-1] == 'a' {
			URL = "https://api.imgur.com/3/album/" + URL[ind+1:]
		} else {
			URL = "https://api.imgur.com/3/image/" + URL[ind+1:]
		}
		rq, err := http.NewRequest("GET", URL, http.NoBody)
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
		return "", &ImgurGalleryEntry{
			ID: payload.Data.ID, Title: payload.Data.Title, Description: payload.Data.Description,
			Link: payload.Data.Description, Images: payload.Data.Images,
		}
	case "i.imgur.com":
		return RESOLVE_FINAL, nil
	}
	return "", nil
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

func (img *ImgurImageEntry) GetInfo() string {
	if img.Title == "" {
		return img.Description
	} else if img.Description == "" {
		return img.Title
	}
	return img.Title + "\n" + img.Description
}

func (img *ImgurImageEntry) GetDimensions() (int, int) { return img.Width, img.Height }

func (img *ImgurImageEntry) GetPostURL() string { return "https://imgur.com/" + img.ID }

func (img *ImgurImageEntry) GetURL() string {
	if img.Mp4 != "" {
		return img.Mp4
	}
	return img.Link
}
