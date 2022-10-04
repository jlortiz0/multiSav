package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type StripQueryResolver byte

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

type BlockingResolver byte

// TODO: Detect when the user's DNS is blocking something and prompt them to switch.
func (BlockingResolver) GetResolvableDomains() []string {
	return nil // []string{"files.catbox.moe"}
}

func (BlockingResolver) ResolveURL(string) (string, ImageEntry) {
	return "", nil
}

func (BlockingResolver) GetRequest(u string) (*http.Response, error) {
	URL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return nil, errors.New("Cannot handle domain " + URL.Host)
}

func (BlockingResolver) Destroy() {}
