package redditapi

import (
	"encoding/json"
	"io"
)

type Submission struct {
	ID              string
	Author          string
	Archived        bool
	Author_fullname string
	Created         Timestamp
	Created_utc     Timestamp
	Clicked         bool
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
	Fullname        string
	reddit          *Reddit
}

func NewSubmission(red *Reddit, id string) *Submission {
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
		return nil
	}
	helper.Data.Children[0].Data.reddit = red
	if helper.Data.Children[0].Data.Created_utc.Time.IsZero() {
		// Why would this be missing???
		helper.Data.Children[0].Data.Created_utc.Time = helper.Data.Children[0].Data.Created.UTC()
	}
	return &helper.Data.Children[0].Data
}

func (sub *Submission) GetComments() {

}

func (sub *Submission) Delete() {

}

func (sub *Submission) Reply() {

}

func (sub *Submission) Edit() {

}

func (sub *Submission) Upvote() {

}

func (sub *Submission) Downvote() {

}

func (sub *Submission) ClearVote() {

}

func (sub *Submission) Report() {

}

func (sub *Submission) Save() error {
	req := sub.reddit.buildRequest("POST", "api/save?id="+sub.Fullname, nilReader)
	_, err := sub.reddit.Client.Do(req)
	if err == nil {
		sub.Saved = true
	}
	return err
}

func (sub *Submission) Unsave() error {
	req := sub.reddit.buildRequest("POST", "api/unsave?id="+sub.Fullname, nilReader)
	_, err := sub.reddit.Client.Do(req)
	if err == nil {
		sub.Saved = false
	}
	return err
}

func (sub *Submission) Hide() {

}

func (sub *Submission) Unhide() {

}
