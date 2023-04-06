package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const UserAgent = "linux:org.jlortiz.multiSav:v0.7.0 (by /u/jlortiz)"

type extType int

const (
	EXT_NONE extType = iota
	EXT_PICTURE
	EXT_VIDEO
	EXT_JXL
)

func getExtType(ext string) extType {
	switch ext {
	case "fmp4":
		fallthrough
	case "mp4":
		fallthrough
	case "webm":
		fallthrough
	case "gif":
		fallthrough
	case "m3u8":
		fallthrough
	case "mov":
		return EXT_VIDEO
	case "png":
		fallthrough
	case "jpg":
		fallthrough
	case "jpeg":
		fallthrough
	case "bmp":
		return EXT_PICTURE
	case "jxl":
		return EXT_JXL
	}
	return EXT_NONE
}

type StripQueryResolver struct{}

func (StripQueryResolver) GetResolvableDomains() []string {
	// Some discord images seem to require the query, but some don't work with it
	// return []string{"media.discordapp.net"}
	return nil
}

func (StripQueryResolver) ResolveURL(s string) (string, ImageEntry) {
	ind := strings.LastIndexByte(s, '?')
	if ind == -1 {
		return RESOLVE_FINAL, nil
	}
	return s[:ind], nil
}

func (StripQueryResolver) GetRequest(u string) (*http.Response, error) {
	return http.DefaultClient.Get(u)
}

type BlockingResolver struct{}

// TODO: Detect when the user's DNS is blocking something and prompt them to switch.
// It seems that catbox.moe likes to lose images... should I add a handler for it?
func (BlockingResolver) GetResolvableDomains() []string {
	return nil // []string{"files.catbox.moe"}
}

func (BlockingResolver) ResolveURL(string) (string, ImageEntry) {
	return RESOLVE_FINAL, nil
}

func (BlockingResolver) GetRequest(u string) (*http.Response, error) {
	URL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return nil, errors.New("Cannot handle domain " + URL.Host)
}

func findByProps(u, p string) (string, error) {
	req, _ := http.NewRequest("GET", u, http.NoBody)
	req.Header.Set("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	s := string(data)
	ind := strings.Index(s, "property=\""+p+"\" ")
	if ind == -1 {
		p += ":url"
		ind = strings.Index(s, "property=\""+p+"\" ")
		if ind == -1 {
			return "", errors.New("property not found")
		}
	}
	s = s[ind:]
	ind = strings.Index(s, "content=\"")
	if ind == -1 {
		return "", errors.New("property has no associated content")
	}
	s = s[ind+len("content=\""):]
	ind = strings.IndexByte(s, '"')
	s = s[:ind]
	return strings.Clone(s), nil
}

type PropOGVideoResolver struct{}

func (PropOGVideoResolver) GetResolvableDomains() []string {
	return []string{}
}

func (PropOGVideoResolver) ResolveURL(u string) (string, ImageEntry) {
	s, _ := findByProps(u, "og:video")
	return s, nil
}

func (PropOGVideoResolver) GetRequest(u string) (*http.Response, error) {
	return http.DefaultClient.Get(u)
}

type PropOGImageResolver struct{}

func (PropOGImageResolver) GetResolvableDomains() []string {
	return []string{"gelbooru.com", "www.gelbooru.com", "danbooru.donmai.us", "ibb.co"}
}

func (PropOGImageResolver) ResolveURL(u string) (string, ImageEntry) {
	s, _ := findByProps(u, "og:image")
	return s, nil
}

func (PropOGImageResolver) GetRequest(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, http.NoBody)
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

type RedgifsResolver struct{}

func (RedgifsResolver) GetResolvableDomains() []string {
	return []string{"redgifs.com", "www.redgifs.com", "*.redgifs.com", "gfycat.com", "www.gfycat.com"}
}

var redgifs_auth = ""

func (RedgifsResolver) ResolveURL(u string) (string, ImageEntry) {
	if strings.Contains(u, "thumbs") {
		return RESOLVE_FINAL, nil
	}
	if strings.Contains(u, "i.redgifs") {
		r, _ := http.Head(u)
		if r == nil || r.StatusCode != 200 {
			return "", nil
		}
		u = r.Request.URL.String()
	}
	var auth struct {
		Token string
	}
	if redgifs_auth == "" {
		req, _ := http.NewRequest("GET", "https://api.redgifs.com/v2/auth/temporary", http.NoBody)
		req.Header.Set("User-Agent", UserAgent)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode/100 != 2 {
			return "", nil
		}
		d := json.NewDecoder(resp.Body)
		err = d.Decode(&auth)
		if err != nil {
			return "", nil
		}
		auth.Token = "Bearer " + auth.Token
	} else {
		auth.Token = redgifs_auth
	}
	ind := strings.LastIndexByte(u, '/')
	req, _ := http.NewRequest("GET", "https://api.redgifs.com/v2/gifs/"+u[ind+1:], http.NoBody)
	req.Header.Set("Authorization", auth.Token)
	req.Header.Set("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode/100 != 2 {
		if resp.StatusCode == 401 && redgifs_auth != "" {
			redgifs_auth = ""
			return RedgifsResolver{}.ResolveURL(u)
		}
		return "", nil
	}
	if redgifs_auth == "" {
		redgifs_auth = auth.Token
	}
	var output struct {
		Gif struct {
			Urls struct {
				HD string
			}
		}
	}
	d := json.NewDecoder(resp.Body)
	d.Decode(&output)
	return output.Gif.Urls.HD, nil
}

func (RedgifsResolver) GetRequest(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, http.NoBody)
	req.Header.Set("User-Agent", UserAgent)
	if redgifs_auth != "" {
		req.Header.Set("Authorization", redgifs_auth)
	}
	return http.DefaultClient.Do(req)
}

type RetryWOQueryResolver struct{}

func (RetryWOQueryResolver) GetResolvableDomains() []string {
	return []string{"cdn.discordapp.com", "media.discordapp.net"}
}

func (RetryWOQueryResolver) ResolveURL(string) (string, ImageEntry) { return "", nil }

func (RetryWOQueryResolver) GetRequest(u string) (*http.Response, error) {
	resp, err := http.Get(u)
	if err != nil || resp.StatusCode/100 == 2 {
		return resp, err
	}
	ind := strings.IndexByte(u, '?')
	if ind == -1 {
		return resp, err
	}
	return http.Get(u[:ind])
}
