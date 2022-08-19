package redditapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
)

type Submission struct {
	ID              string
	Author          string
	Archived        bool
	Author_fullname string
	Created         Timestamp
	Created_utc     Timestamp
	Clicked         bool
	Hidden          bool
	Is_self         bool
	Is_video        bool
	Is_gallery      bool
	Locked          bool
	Num_comments    uint32
	Over_18         bool
	Permalink       string
	Saved           bool
	Score           int
	Selftext        string
	Spoiler         bool
	Stickied        bool
	Subreddit       string
	Title           string
	URL             string
	Upvote_ratio    float32
	Name            string
	Gallery_data    struct {
		Items []struct {
			Media_id string
			// ID int
		}
	}
	Media_metadata map[string]struct {
		M string
		S struct {
			X, Y int
			U    string
			Mp4  string
			Gif  string
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
	reddit                *Reddit
}

func (red *Reddit) Submission(id string) (*Submission, error) {
	var helper struct {
		Data struct {
			Children []struct {
				Data Submission
			}
		}
	}
	req := red.buildRequest("GET", "api/info?id=t3_"+id, nilReader)
	resp, _ := red.Client.Do(req)
	data, _ := io.ReadAll(resp.Body)
	req.Body.Close()
	err := json.Unmarshal(data, &helper)
	if err != nil {
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
	req := sub.reddit.buildRequest("POST", "api/del?id="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := red.buildRequest("POST", fmt.Sprintf("api/comment?thing_id=%s&api_type=json&text=%s", id, url.QueryEscape(text)), nilReader)
	resp, err := red.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", fmt.Sprintf("api/editusertext?thing_id=%s&api_type=json&text=%s", sub.Name, url.QueryEscape(text)), nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := red.buildRequest("POST", fmt.Sprintf("api/vote?id=%s&dir=%d", id, dir), nilReader)
	resp, err := red.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", fmt.Sprintf("api/report?thing_id=%s&reason=%s", sub.Name, reason), nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", "api/save?id="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", "api/unsave?id="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", "api/hide?id="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
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
	req := sub.reddit.buildRequest("POST", "api/unhide?id="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}
