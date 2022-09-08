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
	red := redditapi.NewReddit("linux:org.jlortiz.rediSav:v0.3.2 (by /u/jlortiz)", clientId, clientSecret)
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
			name: "Hot: subreddit",
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
			name: "Rising: subreddit",
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
			name: "Controversial: subreddit",
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
			name: "Top: subreddit",
			args: []ListingArgument{
				{
					name: "Subreddit",
				},
				{
					name: "Best of this",
					options: []interface{}{
						"hour",
						"day",
						"week",
						"month",
						"year",
						"all",
					},
				},
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
					name: "Sort",
					options: []interface{}{
						"relevance",
						"hot",
						"top",
						"new",
						"comments",
					},
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
		{
			name: "Redditor",
			args: []ListingArgument{
				{
					name: "Username",
				},
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
		},
		{
			name: "Personal subreddit",
			args: []ListingArgument{
				{
					name: "Username",
				},
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
			},
			persistent: true,
		},
		{
			name: "Multireddit",
			args: []ListingArgument{
				{
					name: "Username",
				},
				{
					name: "Multi name",
				},
				{
					name: "Track Last",
					kind: LARGTYPE_BOOL,
				},
				// {
				// 	name: "Test int",
				// 	kind: LARGTYPE_INT,
				// },
				// {
				// 	name: "Test url",
				// 	kind: LARGTYPE_URL,
				// },
			},
			persistent: true,
		},
	}
}

type RedditImageListing struct {
	*redditapi.SubmissionIterator
	kind int
	args []interface{}
	seen string
}

func (red *RedditImageListing) GetInfo() (int, []interface{}) {
	return red.kind, red.args
}

func (red *RedditImageListing) GetPersistent() interface{} {
	return red.seen
}

func (red *RedditSite) GetListing(kind int, args []interface{}, persistent interface{}) (ImageListing, []ImageEntry) {
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
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListHot(0)
		}
	case 2:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListRising(0)
		}
	case 3:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListControversial(0)
		}
	case 4:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.ListTop(0, args[1].(string))
		}
	case 5:
		iter, err = red.Self().ListSaved(0)
	case 6:
		var sub *redditapi.Subreddit
		sub, err = red.Subreddit(args[0].(string))
		if err == nil {
			iter, err = sub.Search(0, args[1].(string), "", "")
		}
	case 7:
		iter, err = red.Search(0, args[0].(string), "", "")
	case 8:
		var usr *redditapi.Redditor
		usr, err = red.Redditor(args[0].(string))
		if err == nil {
			iter, err = usr.ListSubmissions(0)
		}
	case 9:
		var usr *redditapi.Redditor
		usr, err = red.Redditor(args[0].(string))
		if err == nil {
			iter, err = usr.UserSubredditListNew(0)
		}
	case 10:
		var multi *redditapi.Multireddit
		multi, err = red.Multireddit(args[0].(string), args[1].(string))
		if err == nil {
			iter, err = multi.ListNew(0)
		}
	}
	if err != nil {
		return &ErrorListing{err}, nil
	}
	if persistent == nil {
		persistent = ""
	}
	iter2 := &RedditImageListing{iter, kind, args, persistent.(string)}
	return iter2, red.ExtendListing(iter2)
}

func (red *RedditSite) ExtendListing(cont ImageListing) []ImageEntry {
	iter2, ok := cont.(*RedditImageListing)
	if !ok {
		return nil
	}
	iter := iter2.SubmissionIterator
	if iter == nil {
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
				// Hack to allow requests to work with crossposts while still loading important fields
				// For instance, fgallery data is only populated for the original post, but just replacing the object can lead to issues saving
				x2 := x.Crosspost_parent_list[len(x.Crosspost_parent_list)-1]
				x.Gallery_data = x2.Gallery_data
				x.Media_metadata = x2.Media_metadata
				x.Preview = x2.Preview
				x.Author = x2.Author
				x.Hidden = x.Hidden || x2.Hidden
				x.Crosspost_parent_list = nil
				x.Is_gallery = x.Is_gallery || x2.Is_gallery
				x.Subreddit = x2.Subreddit
				x.Selftext = x2.Selftext
			}
			if x.Hidden {
				continue
			}
			if strings.HasSuffix(x.URL, ".gifv") {
				// TODO: this doesn't seem to work properly, maybe try substituting .webm instead?
				x.URL = x.URL[:len(x.URL)-1]
			}
			data = append(data, &RedditImageEntry{x})
			if x.Name == iter2.seen {
				iter2.SubmissionIterator = nil
				break
			}
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

func (red *RedditImageEntry) GetGalleryInfo(lazy bool) []ImageEntry {
	if !red.Is_gallery {
		return nil
	}
	if lazy {
		return make([]ImageEntry, len(red.Media_metadata))
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
}

func NewRedditProducer(site *RedditSite, kind int, args []interface{}, persistent interface{}) *RedditProducer {
	return &RedditProducer{NewBufferedImageProducer(site, kind, args, persistent), site}
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
	// TODO: Account for galleries when removing
	// Or should I do this in prodBuf?
	if key == rl.KeyX {
		if red.listing.(*RedditImageListing).kind == 5 {
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
		if red.listing.(*RedditImageListing).kind == 5 {
			useful.Unsave()
		} else {
			useful.Hide()
		}
		red.remove(sel)
		return ARET_MOVEDOWN | ARET_REMOVE
	} else if key == rl.KeyL {
		if red.listing.(*RedditImageListing).kind != 5 {
			args := red.listing.(*RedditImageListing).args
			if args[len(args)-1].(bool) {
				red.listing.(*RedditImageListing).seen = useful.Name
				red.items = red.items[:sel+1]
				for i := BIP_BUFBEFORE + 1; i < BIP_BUFBEFORE+1+BIP_BUFAFTER; i++ {
					red.buffer[i] = nil
				}
				red.listing.(*RedditImageListing).SubmissionIterator = nil
				return ARET_MOVEUP
			}
		}
	}
	return red.BufferedImageProducer.ActionHandler(key, sel, call)
}

func (red *RedditProducer) GetTitle() string {
	useful := red.listing.(*RedditImageListing)
	switch useful.kind {
	case 0:
		return "rediSav - Reddit - New: r/" + useful.args[0].(string)
	case 1:
		return "rediSav - Reddit - Hot: r/" + useful.args[0].(string)
	case 2:
		return "rediSav - Reddit - Rising: r/" + useful.args[0].(string)
	case 3:
		return "rediSav - Reddit - Controversial: r/" + useful.args[0].(string)
	case 4:
		return "rediSav - Reddit - Top: r/" + useful.args[0].(string)
	case 5:
		return "rediSav - Reddit - Saved: u/" + red.site.Self().Name
	case 6:
		return "rediSav - Reddit - Search: r/" + useful.args[0].(string) + " - " + useful.args[1].(string)
	case 7:
		return "rediSav - Reddit - Search - " + useful.args[0].(string)
	case 8:
		return "rediSav - Reddit - New: u/" + useful.args[0].(string)
	case 9:
		return "rediSav - Reddit - New: r/u_" + useful.args[0].(string)
	case 10:
		return "rediSav - Reddit - New: u/" + useful.args[0].(string) + "/m/" + useful.args[1].(string)
	default:
		return "rediSav - Reddit - Unknown"
	}
}
