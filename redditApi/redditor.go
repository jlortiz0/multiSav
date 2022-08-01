package redditapi

import (
	"encoding/json"
	"errors"
	"io"
)

type Redditor struct {
	ID                                                              string
	Is_employee, Is_friend, Is_mod, Is_gold, Is_suspended, Verified bool
	Created_utc                                                     uint64
	Name                                                            string
	Icon_img                                                        string
	Subreddit                                                       string
	Total_karma                                                     int
	reddit                                                          *Reddit
}

func (red *Reddit) Redditor(name string) (*Redditor, error) {
	var helper struct {
		Data Redditor
	}
	rq := red.buildRequest("GET", "user/"+name+"/about", nilReader)
	resp, err := red.Client.Do(rq)
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

func (usr *Redditor) Block() {

}
