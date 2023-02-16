package redditapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Redditor struct {
	ID                                          string
	Is_employee, Is_mod, Is_suspended, Verified bool
	Created_utc                                 Timestamp
	Name                                        string
	Icon_img                                    string
	Subreddit                                   *Subreddit
	Total_karma                                 int
	reddit                                      *Reddit
	self                                        bool
}

func (red *Reddit) Redditor(name string) (*Redditor, error) {
	var helper struct {
		Data Redditor
	}
	rq := red.buildRequest("GET", "user/"+name+"/about", http.NoBody)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(string(data))
	}
	err = json.Unmarshal(data, &helper)
	if err != nil {
		return nil, err
	}
	helper.Data.reddit = red
	return &helper.Data, nil
}

func (usr *Redditor) Block() error {
	rq := usr.reddit.buildRequest("POST", "api/block_user?name="+usr.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (usr *Redditor) Unblock() error {
	rq := usr.reddit.buildRequest("POST", "api/unfriend?type=enemy&name="+usr.Name, http.NoBody)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (usr *Redditor) Report(reason string) error {
	if reason == "" {
		return errors.New("non-empty reason required")
	}
	rq := usr.reddit.buildRequest("POST", "api/report_user?user="+usr.Name+"&reason="+reason, http.NoBody)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	return nil
}

func (usr *Redditor) ListSubmissions(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/submitted", usr.reddit, limit)
}

func (usr *Redditor) ListComments(limit int) (*CommentIterator, error) {
	return newCommentIterator("user/"+usr.Name+"/comments", usr.reddit, limit)
}

func (usr *Redditor) ListDownvoted(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/downvoted", usr.reddit, limit)
}

func (usr *Redditor) ListHidden(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/hidden", usr.reddit, limit)
}

func (usr *Redditor) ListSaved(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/saved", usr.reddit, limit)
}

func (usr *Redditor) ListSavedComments(limit int) (*CommentIterator, error) {
	return newCommentIterator("user/"+usr.Name+"/saved", usr.reddit, limit)
}

func (usr *Redditor) ListUpvoted(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/upvoted", usr.reddit, limit)
}

func (usr *Redditor) ListGilded(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/gilded", usr.reddit, limit)
}

func (usr *Redditor) Multireddits() ([]*Multireddit, error) {
	if usr.self {
		return multiredditSlice("api/multi/mine", usr.reddit)
	}
	return multiredditSlice("api/multi/user/"+usr.Name, usr.reddit)
}

func (usr *Redditor) UserSubredditListNew(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+usr.Name+"/search?q=author:"+usr.Name+"&restrict_sr=true&sort=new", usr.reddit, limit)
}
