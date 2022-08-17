package main

import "jlortiz.org/redisav/redditapi"

type RedditSite struct {
	redditapi.Reddit
}

func NewRedditSite(clientId, clientSecret, user, pass string) *RedditSite {
	red := redditapi.NewReddit("", clientId, clientSecret)
	if user != "" {
		red.Login(user, pass)
	}
	return &RedditSite{*red}
}

func (red *RedditSite) Destroy() {
	red.Logout()
}

func (red *RedditSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			name: "New: subreddit",
			args: []ListingArgument{
				{
					name: "Subreddit",
				},
			},
		},
		{
			name: "New: all",
			args: nil,
		},
	}
}

func (red *RedditSite) GetListing(kind int, args []interface{}) (interface{}, []string) {
	var iter *redditapi.SubmissionIterator
	var err error
	switch kind {
	case 0:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListNew(0)
		}
	case 1:
		iter, err = red.ListNew(0)
	}
	if err != nil {
		return err, nil
	}
	data := make([]string, 0, iter.Buffered())
	for !iter.NextRequiresFetch() {
		x, err := iter.Next()
		if err == nil {
			data = append(data, x.URL)
		}
	}
	return iter, data
}

func (red *RedditSite) ExtendListing(cont interface{}) []string {
	iter, ok := cont.(*redditapi.SubmissionIterator)
	if !ok {
		return nil
	}
	x, err := iter.Next()
	if err != nil || x == nil {
		return nil
	}
	data := make([]string, 1, iter.Buffered()+1)
	data[0] = x.URL
	for !iter.NextRequiresFetch() {
		x, err = iter.Next()
		if err == nil {
			data = append(data, x.URL)
		}
	}
	return data
}
