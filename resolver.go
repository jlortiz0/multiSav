package main

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const UserAgent = "linux:org.jlortiz.multiSav:v0.7.0 (by /u/jlortiz)"

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
	var data [8192]byte
	_, err = io.ReadFull(resp.Body, data[:])
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", err
	}
	s := string(data[:])
	ind := strings.Index(s, "property=\""+p+"\" content=\"")
	if ind == -1 {
		p += ":url"
		ind = strings.Index(s, "property=\""+p+"\" content=\"")
		if ind == -1 {
			return "", errors.New("property not found")
		}
	}
	s = s[ind+len("property=\"\" content=\"")+len(p):]
	ind = strings.IndexByte(s, '"')
	s = s[:ind]
	return strings.Clone(s), nil
}

type PropOGVideoResolver struct{}

func (PropOGVideoResolver) GetResolvableDomains() []string {
	return []string{"gfycat.com", "www.gfycat.com"}
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
	return []string{"gelbooru.com", "www.gelbooru.com", "redgifs.com", "www.redgifs.com", "thumbs4.redgifs.com", "ibb.co"}
}

func (PropOGImageResolver) ResolveURL(u string) (string, ImageEntry) {
	s, _ := findByProps(u, "og:image")
	if s == "" && strings.Contains(u, "redgifs.com") {
		s, _ := findByProps(u, "og:video")
		return s, nil
	}
	return s, nil
}

func (PropOGImageResolver) GetRequest(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, http.NoBody)
	req.Header.Set("User-Agent", UserAgent)
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
