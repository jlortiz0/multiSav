package redditapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Submission struct {
	Created        Timestamp
	Created_utc    Timestamp
	Media_metadata map[string]struct {
		M string
		S struct {
			U    string
			Mp4  string
			Gif  string
			X, Y int
		}
	}
	reddit             *Reddit
	ID                 string
	Author             string
	Author_fullname    string
	Domain             string
	Permalink          string
	Selftext           string
	Subreddit          string
	Title              string
	URL                string
	Name               string
	Unrepliable_reason string
	Gallery_data       struct {
		Items []struct {
			Media_id string
			// ID int
		}
	}
	Preview struct {
		Images []struct {
			ID     string
			Source struct {
				URL           string
				Width, Height int
			}
		}
	}
	Crosspost_parent_list []*Submission
	Score                 int
	Upvote_ratio          float32
	Num_comments          uint32
	Archived              bool
	Clicked               bool
	Hidden                bool
	Is_self               bool
	Is_video              bool
	Is_gallery            bool
	Locked                bool
	Over_18               bool
	Saved                 bool
	Spoiler               bool
	Stickied              bool
	Author_is_blocked     bool
}

func (red *Reddit) Submission(id string) (*Submission, error) {
	var helper struct {
		Data struct {
			Children []struct {
				Data Submission
			}
		}
	}
	req := red.buildRequest("GET", "api/info?id=t3_"+id, http.NoBody)
	resp, _ := http.DefaultClient.Do(req)
	data, _ := io.ReadAll(resp.Body)
	req.Body.Close()
	err := json.Unmarshal(data, &helper)
	if err != nil || len(helper.Data.Children) == 0 {
		return nil, err
	}
	helper.Data.Children[0].Data.reddit = red
	if helper.Data.Children[0].Data.Created_utc.Time.IsZero() {
		// Why would this be missing???
		helper.Data.Children[0].Data.Created_utc.Time = helper.Data.Children[0].Data.Created.UTC()
	}
	return &helper.Data.Children[0].Data, nil
}

func (sub *Submission) Delete() error {
	req := sub.reddit.buildRequest("POST", "api/del?id="+sub.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func replyHelper(red *Reddit, id string, text string) error {
	req := red.buildRequest("POST", fmt.Sprintf("api/comment?thing_id=%s&api_type=json&text=%s", id, url.QueryEscape(text)), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (sub *Submission) Reply(text string) error {
	return replyHelper(sub.reddit, sub.Name, text)
}

func (sub *Submission) Edit(text string) error {
	req := sub.reddit.buildRequest("POST", fmt.Sprintf("api/editusertext?thing_id=%s&api_type=json&text=%s", sub.Name, url.QueryEscape(text)), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func voteHelper(red *Reddit, id string, dir int) error {
	if dir < -1 || dir > 1 {
		return errors.New("dir out of range; expected in [-1, 1]")
	}
	req := red.buildRequest("POST", fmt.Sprintf("api/vote?id=%s&dir=%d", id, dir), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (sub *Submission) Upvote() error {
	return voteHelper(sub.reddit, sub.Name, 1)
}

func (sub *Submission) Downvote() error {
	return voteHelper(sub.reddit, sub.Name, -1)
}

func (sub *Submission) ClearVote() error {
	return voteHelper(sub.reddit, sub.Name, 0)
}

func (sub *Submission) Report(reason string) error {
	if reason == "" {
		return errors.New("non-empty reason required")
	}
	req := sub.reddit.buildRequest("POST", fmt.Sprintf("api/report?thing_id=%s&reason=%s", sub.Name, reason), http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (sub *Submission) Save() error {
	req := sub.reddit.buildRequest("POST", "api/save?id="+sub.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	sub.Saved = true
	return nil
}

func (sub *Submission) Unsave() error {
	req := sub.reddit.buildRequest("POST", "api/unsave?id="+sub.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	sub.Saved = false
	return nil
}

func (sub *Submission) Hide() error {
	req := sub.reddit.buildRequest("POST", "api/hide?id="+sub.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (sub *Submission) Unhide() error {
	req := sub.reddit.buildRequest("POST", "api/unhide?id="+sub.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}
