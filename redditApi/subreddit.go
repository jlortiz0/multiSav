package redditapi

import (
	"encoding/json"
	"errors"
	"io"
)

type Subreddit struct {
	ID                                                    string
	Active_user_count                                     int
	Created_utc                                           uint64
	Description                                           string
	Display_name                                          string
	Name                                                  string
	Over18                                                bool
	Public_description                                    string
	Subscribers                                           int
	User_is_banned, User_is_moderator, User_is_subscriber bool
	reddit                                                *Reddit
}

func NewSubreddit(red *Reddit, id string) *Subreddit {
	var helper struct {
		Data Subreddit
	}
	req := red.buildRequest("GET", "r/"+id+"/about", nilReader)
	resp, _ := red.Client.Do(req)
	data, _ := io.ReadAll(resp.Body)
	req.Body.Close()
	err := json.Unmarshal(data, &helper)
	if err != nil {
		return nil
	}
	helper.Data.reddit = red
	return &helper.Data
}

func (sub *Subreddit) ListNew(limit uint32) (*SubmissionIterator, error) {
	return newSubmissionIterator("r/"+sub.Display_name+"/new", sub.reddit, limit)
}

func (sub *Subreddit) ListHot(limit uint32) (*SubmissionIterator, error) {
	return newSubmissionIterator("r/"+sub.Display_name+"/hot", sub.reddit, limit)
}

func (sub *Subreddit) ListControversial(limit uint32) (*SubmissionIterator, error) {
	return newSubmissionIterator("r/"+sub.Display_name+"/controversial", sub.reddit, limit)
}

func (sub *Subreddit) ListRising(limit uint32) (*SubmissionIterator, error) {
	return newSubmissionIterator("r/"+sub.Display_name+"/rising", sub.reddit, limit)
}

func (sub *Subreddit) ListTop(limit uint32, t string) (*SubmissionIterator, error) {
	s := "r/" + sub.Display_name + "/top"
	if t != "" {
		s += "?t=" + t
	}
	return newSubmissionIterator(s, sub.reddit, limit)
}

func (sub *Subreddit) Search() {

}

func (sub *Subreddit) Subscribe() error {
	req := sub.reddit.buildRequest("POST", "api/subscribe?action=sub&sr="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	sub.User_is_subscriber = true
	return nil
}

func (sub *Subreddit) Unsubscribe() error {
	req := sub.reddit.buildRequest("POST", "api/subscribe?action=unsub&sr="+sub.Name, nilReader)
	resp, err := sub.reddit.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(string(data))
	}
	sub.User_is_subscriber = false
	return nil
}
