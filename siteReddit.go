package main

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
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
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
		},
		{
			name: "New: all",
			args: []ListingArgument{
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
		},
		{
			name: "Saved",
		},
		{
			name: "Search: subreddit",
			args: []ListingArgument{
				{
					name: "Subreddit",
				},
				{
					name: "Search",
				},
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
		},
		{
			name: "Search: all",
			args: []ListingArgument{
				{
					name: "Search",
				},
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
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
	case 2:
		iter, err = red.Self().ListSaved(0)
	case 3:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.Search(0, args[1].(string), "", "")
		}
	case 4:
		// Need to add reddit.Search()
	}
	if err != nil {
		return err, nil
	}
	return iter, red.ExtendListing(iter)
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
	data[0] = &RedditImageEntry{x}
	for !iter.NextRequiresFetch() {
		x, err = iter.Next()
		if err == nil {
			if len(x.Crosspost_parent_list) != 0 {
				x = x.Crosspost_parent_list[len(x.Crosspost_parent_list)-1]
			}
			if x.Hidden {
				continue
			}
			if strings.HasSuffix(x.URL, ".gifv") {
				x.URL = x.URL[:len(x.URL)-1]
			}
			data = append(data, &RedditImageEntry{x})
		}
	}
	return data
}

type RedditImageEntry struct {
	*redditapi.Submission
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
	for i, s := range red.Gallery_data.Items {
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
		data = append(data, &RedditGalleryEntry{RedditImageEntry: *red, name: fmt.Sprintf("%s (%d/%d)", red.Title, i+1, len(red.Gallery_data.Items)), url: x.S.U, x: x.S.X, y: x.S.Y})
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

const RIE_LINE_BREAK_CHARS = 100
const RIE_LINE_BREAK_TOLERANCE = 10

func wordWrapper(s string) string {
	s2 := new(strings.Builder)
	for len(s) > RIE_LINE_BREAK_CHARS+RIE_LINE_BREAK_TOLERANCE {
		ind := strings.IndexByte(s, '\n')
		if ind != -1 && ind < RIE_LINE_BREAK_CHARS {
			s2.WriteString(s[:ind+1])
			s = s[ind+1:]
			continue
		}
		ind = strings.IndexAny(s[RIE_LINE_BREAK_CHARS-RIE_LINE_BREAK_TOLERANCE:RIE_LINE_BREAK_CHARS], " \t-\r")
		if ind != -1 {
			s2.WriteString(s[:RIE_LINE_BREAK_CHARS-RIE_LINE_BREAK_TOLERANCE+ind])
			s2.WriteByte('\n')
			s = s[RIE_LINE_BREAK_CHARS-RIE_LINE_BREAK_TOLERANCE+ind:]
			if s[0] != '-' {
				s = s[1:]
			}
		} else {
			ind = strings.IndexAny(s[RIE_LINE_BREAK_CHARS:RIE_LINE_BREAK_CHARS+RIE_LINE_BREAK_TOLERANCE], " \t-\r")
			if ind != -1 {
				s2.WriteString(s[:RIE_LINE_BREAK_CHARS+ind])
				s2.WriteByte('\n')
				s = s[RIE_LINE_BREAK_CHARS+ind:]
				if s[0] != '-' {
					s = s[1:]
				}
			} else {
				s2.WriteString(s[:RIE_LINE_BREAK_CHARS])
				s2.WriteByte('\n')
				s = s[RIE_LINE_BREAK_CHARS:]
			}
		}
	}
	s2.WriteString(s)
	return s2.String()
}

func (red *RedditImageEntry) GetText() string {
	return wordWrapper(red.Selftext)
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

func (red *RedditImageEntry) GetPostURL() string {
	return "https://reddit.com" + red.Permalink
}

func (red *RedditImageEntry) GetInfo() string {
	return fmt.Sprintf("%s by u/%s (r/%s)\nScore: %d\nComments: %d", red.Title, red.Author, red.Subreddit, red.Score, red.Num_comments)
}

type RedditGalleryEntry struct {
	RedditImageEntry
	name string
	url  string
	x, y int
}

func (red *RedditGalleryEntry) GetURL() string {
	return red.url
}

func (red *RedditGalleryEntry) GetDimensions() (int, int) {
	return red.x, red.y
}

func (red *RedditGalleryEntry) GetName() string {
	return red.name
}

func (*RedditGalleryEntry) GetType() ImageEntryType {
	return IETYPE_REGULAR
}

type RedditProducer struct {
	*BufferedImageProducer
	site *RedditSite
	kind int
	args []interface{}
}

func NewRedditProducer(site *RedditSite, kind int, args []interface{}) *RedditProducer {
	return &RedditProducer{NewBufferedImageProducer(site, kind, args), site, kind, args}
}

func (red *RedditProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	var useful *redditapi.Submission
	switch v := red.items[sel].(type) {
	case *RedditImageEntry:
		useful = v.Submission
	case *RedditGalleryEntry:
		useful = v.Submission
	default:
		return red.BufferedImageProducer.ActionHandler(key, sel, call)
	}
	if key == rl.KeyX {
		if red.kind == 2 {
			ret := red.BufferedImageProducer.ActionHandler(key, sel, call)
			if ret&ARET_REMOVE != 0 {
				useful.Unsave()
			}
			return ret
		} else {
			useful.Save()
		}
		red.remove(sel)
		return ARET_MOVEUP | ARET_REMOVE
	} else if key == rl.KeyC {
		if red.kind == 2 {
			useful.Unsave()
		} else {
			useful.Hide()
		}
		red.remove(sel)
		return ARET_MOVEDOWN | ARET_REMOVE
	}
	return red.BufferedImageProducer.ActionHandler(key, sel, call)
}
