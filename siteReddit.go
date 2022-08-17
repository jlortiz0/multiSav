package main

import (
	"strings"

	"jlortiz.org/redisav/redditapi"
)

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

func (red *RedditSite) GetListing(kind int, args []interface{}) (interface{}, []ImageEntry) {
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
	data := make([]ImageEntry, 0, iter.Buffered())
	for !iter.NextRequiresFetch() {
		x, err := iter.Next()
		if err == nil {
			if len(x.Crosspost_parent_list) != 0 {
				x = x.Crosspost_parent_list[len(x.Crosspost_parent_list)-1]
			}
			data = append(data, &RedditImageEntry{*x})
		}
	}
	return iter, data
}

func (red *RedditSite) ExtendListing(cont interface{}) []ImageEntry {
	iter, ok := cont.(*redditapi.SubmissionIterator)
	if !ok {
		return nil
	}
	x, err := iter.Next()
	if err != nil || x == nil {
		return nil
	}
	data := make([]ImageEntry, 1, iter.Buffered()+1)
	data[0] = &RedditImageEntry{*x}
	for !iter.NextRequiresFetch() {
		x, err = iter.Next()
		if err == nil {
			if len(x.Crosspost_parent_list) != 0 {
				x = x.Crosspost_parent_list[len(x.Crosspost_parent_list)-1]
			}
			data = append(data, &RedditImageEntry{*x})
		}
	}
	return data
}

type RedditImageEntry struct {
	redditapi.Submission
}

func (red *RedditImageEntry) GetType() ImageEntryType {
	if red.Is_self {
		return IETYPE_TEXT
	}
	if red.Is_gallery {
		return IETYPE_GALLERY
	}
	if red.Is_video {
		return IETYPE_ANIMATED
	}
	return IETYPE_REGULAR
}

func (red *RedditImageEntry) GetGalleryInfo() []ImageEntry {
	if !red.Is_gallery {
		return nil
	}
	data := make([]ImageEntry, 0, len(red.Media_metadata))
	for _, s := range red.Gallery_data.Items {
		x := red.Media_metadata[s.Media_id]
		if x.S.U == "" {
			x.S.U = x.S.Mp4
		}
		ind := strings.LastIndexByte(x.S.U, '/')
		if ind != -1 {
			x.S.U = "https://i.redd.it" + x.S.U[ind:]
		}
		ind = strings.IndexByte(x.S.U, '?')
		if ind != -1 {
			x.S.U = x.S.U[:ind]
		}
		data = append(data, &DummyImageEntry{name: "whatever", url: x.S.U, kind: IETYPE_REGULAR, x: x.S.X, y: x.S.Y})
	}
	return data
}

func (red *RedditImageEntry) GetName() string {
	return red.Title
}

func (red *RedditImageEntry) GetURL() string {
	if red.Is_self || red.Is_gallery {
		return "https://reddit.com" + red.Permalink
	}
	return red.URL
}

func (red *RedditImageEntry) GetText() string {
	return red.Selftext
}

func (red *RedditImageEntry) GetDimensions() (int, int) {
	if len(red.Preview.Images) != 0 {
		return red.Preview.Images[0].Source.Width, red.Preview.Images[0].Source.Height
	}
	if len(red.Media_metadata) != 0 {
		for _, v := range red.Media_metadata {
			return v.S.X, v.S.Y
		}
	}
	return 0, 0
}
