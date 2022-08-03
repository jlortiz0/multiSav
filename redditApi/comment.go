package redditapi

import (
	"errors"
	"fmt"
	"io"
	"net/url"
)

type Comment struct {
	Subreddit       string
	ID              string
	Name            string
	Saved           bool
	Created_utc     Timestamp
	Author_fullname string
	Score           int
	Body            string
	Parent_id       string
	Link_id         string
	Is_submitter    bool
	Removed         bool
	Edited          bool
	reddit          *Reddit
}

func (c *Comment) Reply(text string) error {
	return replyHelper(c.reddit, c.Name, text)
}

func (c *Comment) Upvote() error {
	return voteHelper(c.reddit, c.Name, 1)
}

func (c *Comment) Downvote() error {
	return voteHelper(c.reddit, c.Name, -1)
}

func (c *Comment) ClearVote() error {
	return voteHelper(c.reddit, c.Name, 0)
}

func (c *Comment) Save() error {
	req := c.reddit.buildRequest("POST", "api/save?id="+c.Name, nilReader)
	resp, err := c.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	c.Saved = true
	return nil
}

func (c *Comment) Unsave() error {
	req := c.reddit.buildRequest("POST", "api/unsave?id="+c.Name, nilReader)
	resp, err := c.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	c.Saved = false
	return nil
}

func (c *Comment) Edit(text string) error {
	req := c.reddit.buildRequest("POST", fmt.Sprintf("api/editusertext?thing_id=%s&api_type=json&text=%s", c.Name, url.QueryEscape(text)), nilReader)
	resp, err := c.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (c *Comment) Delete() error {
	req := c.reddit.buildRequest("POST", "api/del?id="+c.Name, nilReader)
	resp, err := c.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (c *Comment) Report(reason string) error {
	if reason == "" {
		return errors.New("non-empty reason required")
	}
	req := c.reddit.buildRequest("POST", fmt.Sprintf("api/report?thing_id=%s&reason=%s", c.Name, reason), nilReader)
	resp, err := c.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}
