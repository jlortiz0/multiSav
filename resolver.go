package main

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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

func (StripQueryResolver) Destroy() {}

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

func (BlockingResolver) Destroy() {}

type GfycatResolver struct{}

func (GfycatResolver) GetResolvableDomains() []string {
	return []string{"gfycat.com", "www.gfycat.com", "thumbs.gfycat.com"}
}

func (GfycatResolver) ResolveURL(u string) (string, ImageEntry) {
	if strings.HasPrefix(u, "thumbs.") {
		return RESOLVE_FINAL, nil
	}
	resp, err := http.Get(u)
	if err != nil {
		return "", nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}
	s := string(data)
	ind := strings.Index(s, "property=\"og:video\" content=\"")
	if ind == -1 {
		return "", nil
	}
	s = s[ind+len("property=\"og:video\" content=\""):]
	ind = strings.IndexByte(s, '"')
	s = s[:ind]
	return s, nil
}

func (GfycatResolver) GetRequest(u string) (*http.Response, error) {
	return http.DefaultClient.Get(u)
}

func (GfycatResolver) Destroy() {}
