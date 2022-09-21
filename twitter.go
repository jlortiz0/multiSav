package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// TODO: Turns out that I don't have access to v1.1
// My options are use another library that uses the v2 endpoints or write my own
// I'm assuming that the user might not be able to access v1.1 in the future (since I'm a cheapskate and will make them provide thier own token)
// (at least with reddit i have the excuse that I don't have an authentication website so there's no way for the user to log in otherwise)
type TwitterSite struct {
	*twitter.Client
	client *http.Client
}

func NewTwitterSite(id, secret string) TwitterSite {
	auth := &clientcredentials.Config{
		ClientID:     id,
		ClientSecret: secret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	c := auth.Client(oauth2.NoContext)
	return TwitterSite{twitter.NewClient(c), c}
}

func (t TwitterSite) Destroy() {}

func (t TwitterSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			"User timeline", []ListingArgument{{name: "User"}, {name: "Include retweets", kind: LARGTYPE_BOOL}, {name: "Include replies", kind: LARGTYPE_BOOL}},
		},
		{
			"Search: 7 days", []ListingArgument{{name: "Query"}, {name: "Sort", options: []interface{}{"mixed", "recent", "popular"}}},
		},
		{
			"Search: user, 7 days", []ListingArgument{{name: "User"}, {name: "Query"}, {name: "Sort", options: []interface{}{"mixed", "recent", "popular"}}},
		},
		{
			"List", []ListingArgument{{name: "ID or URL"}, {name: "Include retweets", kind: LARGTYPE_BOOL}},
		},
		{
			"Home", []ListingArgument{{name: "Include replies", kind: LARGTYPE_BOOL}},
		},
	}
}

type TwitterImageListing struct {
	site    TwitterSite
	kind    int
	args    []interface{}
	lastId  int64
	persist int64
}

func (t *TwitterImageListing) GetInfo() (int, []interface{}) {
	return t.kind, t.args
}

func (t *TwitterImageListing) GetPersistent() interface{} {
	return t.persist
}

func (t TwitterSite) GetListing(kind int, args []interface{}, persist interface{}) (ImageListing, []ImageEntry) {
	var tweets []twitter.Tweet
	var err error
	var stopAt int64
	if persist != nil {
		stopAt = persist.(int64)
	}
	switch kind {
	case 0:
		// User timeline
		includeRts := args[1].(bool)
		includeRpl := !args[2].(bool)
		tweets, _, err = t.Timelines.UserTimeline(&twitter.UserTimelineParams{Count: 100, ScreenName: args[0].(string), ExcludeReplies: &includeRpl, IncludeRetweets: &includeRts, SinceID: stopAt})
	case 1:
		// Search
		var search *twitter.Search
		search, _, err = t.Search.Tweets(&twitter.SearchTweetParams{Query: args[0].(string), Count: 100, ResultType: args[1].(string), SinceID: stopAt})
		tweets = search.Statuses
	case 2:
		// Search user
		nArgs := args[1:]
		nArgs[0] = args[1].(string) + " from:" + args[0].(string)
		return t.GetListing(1, nArgs, persist)
	case 3:
		// List
		s := args[0].(string)
		ind := strings.LastIndexByte(s, '/')
		if ind != -1 {
			s = s[ind+1:]
		}
		args[0], err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			break
		}
		includeRts := args[1].(bool)
		tweets, _, err = t.Lists.Statuses(&twitter.ListsStatusesParams{ListID: args[0].(int64), Count: 100, IncludeRetweets: &includeRts, SinceID: stopAt})
	case 4:
		// Home
		includeRpl := !args[0].(bool)
		tweets, _, err = t.Timelines.HomeTimeline(&twitter.HomeTimelineParams{Count: 100, SinceID: stopAt, ExcludeReplies: &includeRpl})
	default:
		err = errors.New("unknown type")
	}
	if err != nil {
		return ErrorListing{err}, nil
	}
	thing := new(TwitterImageListing)
	thing.site = t
	thing.persist = stopAt
	thing.kind = kind
	thing.args = args
	thing.lastId = tweets[len(tweets)-1].ID
	nTweets := make([]ImageEntry, len(tweets))
	for i, x := range tweets {
		nTweets[i] = &TwitterImageEntry{x.ID, x.FullText, x.ExtendedEntities.Media, x.Entities.Urls, x.FavoriteCount, x.ReplyCount, x.RetweetCount, x.QuoteCount, x.User.Name}
	}
	return thing, nTweets
}

func (t TwitterSite) ExtendListing(ls ImageListing) []ImageEntry {
	thing, ok := ls.(*TwitterImageListing)
	if !ok {
		return nil
	}
	args := thing.args
	var tweets []twitter.Tweet
	var err error
	stopAt := thing.persist
	switch thing.kind {
	case 0:
		// User timeline
		includeRts := !args[1].(bool)
		includeRpl := args[2].(bool)
		tweets, _, err = t.Timelines.UserTimeline(&twitter.UserTimelineParams{Count: 100, ScreenName: args[0].(string), ExcludeReplies: &includeRpl, IncludeRetweets: &includeRts, SinceID: stopAt, MaxID: thing.lastId})
	// case 2:
	// 	// This code should never run
	// 	s := thing.args[1].(string)
	// 	s += " from: " + thing.args[0].(string)
	// 	thing.args[1] = s
	// 	thing.args = thing.args[1:]
	// 	args = thing.args
	// 	thing.kind = 2
	// 	fallthrough
	case 1:
		// Search
		var search *twitter.Search
		search, _, err = t.Search.Tweets(&twitter.SearchTweetParams{Query: args[0].(string), Count: 100, ResultType: args[1].(string), SinceID: stopAt, MaxID: thing.lastId})
		tweets = search.Statuses
	case 3:
		// List
		includeRts := args[1].(bool)
		tweets, _, err = t.Lists.Statuses(&twitter.ListsStatusesParams{ListID: args[0].(int64), Count: 100, IncludeRetweets: &includeRts, SinceID: stopAt, MaxID: thing.lastId})
	case 4:
		// Home
		includeRpl := !args[0].(bool)
		tweets, _, err = t.Timelines.HomeTimeline(&twitter.HomeTimelineParams{Count: 100, SinceID: stopAt, ExcludeReplies: &includeRpl, MaxID: thing.lastId})
	default:
		err = strconv.ErrRange
	}
	if err != nil {
		return nil
	}
	nTweets := make([]ImageEntry, len(tweets))
	for i, x := range tweets {
		nTweets[i] = &TwitterImageEntry{x.ID, x.FullText, x.ExtendedEntities.Media, x.Entities.Urls, x.FavoriteCount, x.ReplyCount, x.RetweetCount, x.QuoteCount, x.User.Name}
		if x.ID == thing.persist {
			break
		}
	}
	return nTweets
}

func (t TwitterSite) GetRequest(URL string) (*http.Response, error) {
	return t.client.Get(URL)
}

func (t TwitterSite) ResolveURL(URL string) (string, ImageEntry) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", nil
	}
	switch u.Hostname() {
	case "www.twitter.com":
		fallthrough
	case "twitter.com":
		s := u.Path
		ind := strings.Index(s, "status")
		if ind == -1 {
			return "", nil
		}
		s = s[ind+6:]
		id, _ := strconv.ParseInt(s, 10, 64)
		if id == 0 {
			return "", nil
		}
		trueConst := true
		x, _, err := t.Statuses.Show(id, &twitter.StatusShowParams{ID: id, IncludeEntities: &trueConst})
		if err != nil {
			return "", nil
		}
		return "", &TwitterImageEntry{x.ID, x.FullText, x.ExtendedEntities.Media, x.Entities.Urls, x.FavoriteCount, x.ReplyCount, x.RetweetCount, x.QuoteCount, x.User.Name}
	case "pbs.twimg.com":
		fallthrough
	case "twimg.com":
		return RESOLVE_FINAL, nil
	}
	return "", nil
}

func (t TwitterSite) GetResolvableDomains() []string {
	return []string{"twitter.com", "pbs.twimg.com", "twimg.com", "www.twitter.com"}
}

type TwitterImageEntry struct {
	ID                 int64
	text               string
	media              []twitter.MediaEntity
	url                []twitter.URLEntity
	fav, repl, rt, qrt int
	username           string
}

func (t *TwitterImageEntry) GetName() string {
	return t.username
}

func (t *TwitterImageEntry) GetPostURL() string {
	return fmt.Sprintf("https://twitter.com/%s/status/%d", t.username, t.ID)
}

func (t *TwitterImageEntry) GetText() string {
	return wordWrapper(t.text)
}

func (t *TwitterImageEntry) GetDimensions() (int, int) {
	if len(t.media) == 0 {
		return -1, -1
	}
	return t.media[0].Sizes.Large.Width, t.media[0].Sizes.Large.Height
}

func (t *TwitterImageEntry) GetInfo() string {
	return fmt.Sprintf("%s\n@%s\nL: %d  RP: %d  RT: %d  QRT: %d", t.text, t.username, t.fav, t.repl, t.rt, t.qrt)
}

func (t *TwitterImageEntry) GetType() ImageEntryType {
	if len(t.media) > 1 {
		return IETYPE_GALLERY
	} else if len(t.media) > 0 {
		return IETYPE_REGULAR
		// } else if len(t.url) > 0 {
		// 	return IETYPE_REGULAR
	} else {
		return IETYPE_TEXT
	}
}

func (t *TwitterImageEntry) GetURL() string {
	if len(t.media) > 0 {
		s := t.media[0].MediaURL
		ind := strings.LastIndexByte(s, '.')
		return fmt.Sprintf("%s?format=%s&name=large", s[:ind], s[ind+1:])
	} else if len(t.url) > 0 {
		return t.url[0].ExpandedURL
	} else {
		return t.GetPostURL()
	}
}

func (*TwitterImageEntry) Combine(ImageEntry) {}

func (t *TwitterImageEntry) GetGalleryInfo(b bool) []ImageEntry {
	imgs := make([]ImageEntry, len(t.media))
	if !b {
		return imgs
	}
	for i, x := range t.media {
		imgs[i] = &TwitterGalleryEntry{*t, x.MediaURL, x.Sizes.Large.Width, x.Sizes.Large.Height}
	}
	return imgs
}

type TwitterGalleryEntry struct {
	TwitterImageEntry
	url  string
	w, h int
}

func (t *TwitterGalleryEntry) GetURL() string {
	return t.url
}

func (t *TwitterGalleryEntry) GetDimensions() (int, int) {
	return t.w, t.h
}
