package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"os"
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
	req := p.buildGetRequest(URL)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug_output {
		f, _ := os.OpenFile("out.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		f.Write(data)
		f.Close()
	}
	var output struct {
		Illusts  []*Illustration
		Next_url string
		Error    struct {
			Message string
		}
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
		new, err := ls.client.newIllustrationListing(ls.nextURL)
		if err != nil || new == nil || len(new.data) == 0 {
			ls.lazy = false
			return nil, err
		}
		// ls.count += new.count
		ls.nextURL = new.nextURL
		ls.index = 0
		ls.data = new.data
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
	req := p.buildGetRequest(URL)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug_output {
		f, _ := os.OpenFile("out.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		f.Write(data)
		f.Close()
	}
	var output struct {
		User_previews []struct {
			User *User
		}
		Next_url string
		Error    struct {
			Message string
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
		new, err := ls.client.newUserListing(ls.nextURL)
		if err != nil || new == nil || len(new.data) == 0 {
			ls.lazy = false
			return nil, err
		}
		// ls.count += new.count
		ls.nextURL = new.nextURL
		ls.index = 0
		ls.data = new.data
	}
	ls.index++
	ls.count++
	ls.data[ls.index-1].client = ls.client
	return ls.data[ls.index-1], nil
}
