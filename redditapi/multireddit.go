package redditapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Multireddit struct {
	Name            string
	Display_name    string
	Num_subscribers int
	Copied_from     string
	Subreddits      []struct {
		Name string
	}
	Created_utc Timestamp
	Visibility  string
	Over_18     bool
	Path        string
	Owner       string
	reddit      *Reddit
}

func (red *Reddit) Multireddit(user string, name string) (*Multireddit, error) {
	var helper struct {
		Data Multireddit
	}
	rq := red.buildRequest("GET", "api/multi/user/"+user+"/m/"+name, http.NoBody)
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

func multiredditSlice(url string, red *Reddit) ([]*Multireddit, error) {
	req := red.buildRequest("GET", url, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(string(data))
	}
	var payload []struct {
		Data *Multireddit
	}
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	output := make([]*Multireddit, len(payload))
	for i, v := range payload {
		v.Data.reddit = red
		output[i] = v.Data
	}
	return output, nil
}

func (multi *Multireddit) ListNew(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+multi.Owner+"/m/"+multi.Name+"/new", multi.reddit, limit)
}

func (multi *Multireddit) ListHot(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+multi.Owner+"/m/"+multi.Name+"/hot", multi.reddit, limit)
}

func (multi *Multireddit) ListControversial(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+multi.Owner+"/m/"+multi.Name+"/controversial", multi.reddit, limit)
}

func (multi *Multireddit) ListRising(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("user/"+multi.Owner+"/m/"+multi.Name+"/rising", multi.reddit, limit)
}

// If t not specified, seems to default to "day"
func (multi *Multireddit) ListTop(limit int, t string) (*SubmissionIterator, error) {
	s := "user/" + multi.Owner + "/m/" + multi.Name + "/top"
	if t != "" {
		s += "?t=" + t
	}
	return newSubmissionIterator(s, multi.reddit, limit)
}
