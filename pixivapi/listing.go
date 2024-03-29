package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

type IllustrationListing struct {
	client  *Client
	nextURL string
	data    []*Illustration
	index   int
	count   int
	// lastId  int
	lazy bool
}

func (p *Client) newIllustrationListing(URL string) (*IllustrationListing, error) {
	if !strings.HasPrefix(URL, base_url) {
		URL = base_url + URL
	}
	resp, err := p.GetRequest(URL)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output struct {
		Next_url string
		Error    struct {
			Message string
		}
		Illusts []*Illustration
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	ls := new(IllustrationListing)
	ls.client = p
	ls.data = output.Illusts
	// for _, x := range ls.data {
	// 	x.client = p
	// }
	ls.nextURL = output.Next_url
	// ls.count = len(ls.data)
	// ls.lastId = ls.data[len(ls.data)-1].ID
	ls.lazy = len(ls.data) != 0
	return ls, nil
}

func (ls *IllustrationListing) HasNext() bool {
	return ls.lazy || ls.index < len(ls.data)
}

func (ls *IllustrationListing) NextRequiresFetch() bool {
	return ls.index == len(ls.data)
}

func (ls *IllustrationListing) Buffered() int {
	return len(ls.data) - ls.index
}

func (ls *IllustrationListing) Next() (*Illustration, error) {
	if !ls.HasNext() {
		return nil, nil
	}
	if ls.NextRequiresFetch() {
		newLs, err := ls.client.newIllustrationListing(ls.nextURL)
		if err != nil || newLs == nil || len(newLs.data) == 0 {
			ls.lazy = false
			return nil, err
		}
		// ls.count += new.count
		ls.nextURL = newLs.nextURL
		ls.index = 0
		ls.data = newLs.data
	}
	ls.index++
	ls.count++
	ls.data[ls.index-1].client = ls.client
	return ls.data[ls.index-1], nil
}

type UserListing struct {
	client  *Client
	nextURL string
	data    []*User
	index   int
	count   int
	// lastId  int
	lazy bool
}

func (p *Client) newUserListing(URL string) (*UserListing, error) {
	if !strings.HasPrefix(URL, base_url) {
		URL = base_url + URL
	}
	resp, err := p.GetRequest(URL)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output struct {
		Next_url string
		Error    struct {
			Message string
		}
		User_previews []struct {
			User *User
		}
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	ls := new(UserListing)
	ls.client = p
	ls.data = make([]*User, len(output.User_previews))
	for i, x := range output.User_previews {
		ls.data[i] = x.User
	}
	ls.nextURL = output.Next_url
	// ls.count = len(ls.data)
	// ls.lastId = ls.data[len(ls.data)-1].ID
	ls.lazy = len(ls.data) != 0
	return ls, nil
}

func (ls *UserListing) HasNext() bool {
	return ls.lazy || ls.index < len(ls.data)
}

func (ls *UserListing) NextRequiresFetch() bool {
	return ls.index == len(ls.data)
}

func (ls *UserListing) Buffered() int {
	return len(ls.data) - ls.index
}

func (ls *UserListing) Next() (*User, error) {
	if !ls.HasNext() {
		return nil, nil
	}
	if ls.NextRequiresFetch() {
		newLs, err := ls.client.newUserListing(ls.nextURL)
		if err != nil || newLs == nil || len(newLs.data) == 0 {
			ls.lazy = false
			return nil, err
		}
		// ls.count += new.count
		ls.nextURL = newLs.nextURL
		ls.index = 0
		ls.data = newLs.data
	}
	ls.index++
	ls.count++
	ls.data[ls.index-1].client = ls.client
	return ls.data[ls.index-1], nil
}
