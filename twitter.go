package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

type TwitterSite struct {
	*twitter.Client
}

type twitterAuthorizer struct {
	key string
}

func (t twitterAuthorizer) Add(req *http.Request) {
	req.Header.Add("Authorization", t.key)
}

func NewTwitterSite(bearer string) TwitterSite {
	if !strings.HasPrefix(bearer, "Bearer ") {
		bearer = "Bearer " + bearer
	}
	return TwitterSite{&twitter.Client{Authorizer: twitterAuthorizer{bearer}, Client: http.DefaultClient, Host: "https://api.twitter.com"}}
}

func (t TwitterSite) Destroy() {}

func (t TwitterSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			"User timeline", []ListingArgument{{name: "User"}, {name: "Include retweets", kind: LARGTYPE_BOOL}},
		},
		{
			"Search: 7 days", []ListingArgument{{name: "Query"}, {name: "Sort", options: []interface{}{twitter.TweetSearchSortOrderRecency, twitter.TweetSearchSortOrderRelevancy}}},
		},
		{
			"Search: user, 7 days", []ListingArgument{{name: "User"}, {name: "Query"}, {name: "Sort", options: []interface{}{twitter.TweetSearchSortOrderRecency, twitter.TweetSearchSortOrderRelevancy}}},
		},
		{
			"List", []ListingArgument{{name: "ID or URL"}, {name: "Include retweets and replies", kind: LARGTYPE_BOOL}},
		},
		{
			"Home", []ListingArgument{{name: "Include retweets", kind: LARGTYPE_BOOL}},
		},
		{
			"Bookmarks", nil,
		},
	}
}

type TwitterImageListing struct {
	site    TwitterSite
	kind    int
	args    []interface{}
	token   string
	persist string
}

func (t *TwitterImageListing) GetInfo() (int, []interface{}) {
	return t.kind, t.args
}

func (t *TwitterImageListing) GetPersistent() interface{} {
	return t.persist
}

func (t TwitterSite) getMyId() (string, error) {
	req, _ := http.NewRequest("GET", t.Host+"/2/users/me", http.NoBody)
	t.Authorizer.Add(req)
	resp, err := t.Client.Client.Do(req)
	if err != nil {
		return "", err
	}
	var temp struct {
		Data struct {
			ID string
		}
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(data, &temp)
	return temp.Data.ID, err
}

func (t TwitterSite) GetListing(kind int, args []interface{}, persist interface{}) (ImageListing, []ImageEntry) {
	thing := new(TwitterImageListing)
	thing.site = t
	if persist != nil {
		thing.persist = persist.(string)
	}
	if kind == 2 {
		// Search user
		nArgs := args[1:]
		nArgs[0] = args[1].(string) + " from:" + args[0].(string)
		args = nArgs
		kind = 1
	}
	thing.kind = kind
	thing.args = args
	return thing, t.ExtendListing(thing)
}

func (t TwitterSite) ExtendListing(ls ImageListing) []ImageEntry {
	thing, ok := ls.(*TwitterImageListing)
	if !ok {
		return nil
	}
	args := thing.args
	var tweets *twitter.TweetRaw
	var err error
	stopAt := thing.persist
	token := thing.token
	switch thing.kind {
	case 0:
		// User timeline
		var user *twitter.UserLookupResponse
		user, err = t.UserNameLookup(context.Background(), []string{args[0].(string)}, twitter.UserLookupOpts{})
		if err != nil {
			break
		}
		if len(user.Raw.Errors) != 0 {
			err = errors.New(user.Raw.Errors[0].Title)
			break
		}
		if len(user.Raw.Users) == 0 {
			err = errors.New("No such user " + args[0].(string))
			break
		}
		excludes := make([]twitter.Exclude, 1, 2)
		excludes[0] = twitter.ExcludeReplies
		if !args[1].(bool) {
			excludes = append(excludes, twitter.ExcludeRetweets)
		}
		var resp *twitter.UserTweetTimelineResponse
		resp, err = t.UserTweetTimeline(context.Background(), user.Raw.Users[0].ID, twitter.UserTweetTimelineOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields:     []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
			MaxResults:      100,
			Excludes:        excludes,
			SinceID:         stopAt,
			PaginationToken: token,
		})
		if resp == nil {
			break
		}
		tweets = resp.Raw
		token = resp.Meta.NextToken
		if !args[1].(bool) {
			for i, x := range tweets.Tweets {
				if len(x.ReferencedTweets) != 0 {
					tweets.Tweets[i] = nil
				}
			}
		}
	case 1:
		// Search
		var search *twitter.TweetSearchResponse
		search, err = t.TweetSearch(context.Background(), args[0].(string), twitter.TweetSearchOpts{
			Expansions:  []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields: []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields: []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
			MaxResults:  100,
			SinceID:     stopAt,
			SortOrder:   args[1].(twitter.TweetSearchSortOrder),
			NextToken:   token,
		})
		if err != nil {
			break
		}
		if len(search.Raw.Errors) != 0 {
			err = errors.New(search.Raw.Errors[0].Title)
			break
		}
		tweets = search.Raw
		token = search.Meta.NextToken
	case 3:
		// List
		s := args[0].(string)
		var resp *twitter.ListTweetLookupResponse
		resp, err = t.ListTweetLookup(context.Background(), s, twitter.ListTweetLookupOpts{
			Expansions: []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			// TODO: ?!? this is part of the request? Maybe submit a PR to the package to fix this.
			// MediaFields: []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields:     []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
			MaxResults:      100,
			PaginationToken: token,
		})
		if err != nil {
			break
		}
		if len(resp.Raw.Errors) != 0 {
			err = errors.New(resp.Raw.Errors[0].Title)
			break
		}
		tweets = resp.Raw
		for i, x := range tweets.Tweets {
			if x.ID == stopAt {
				tweets.Tweets = tweets.Tweets[:i]
				break
			}
			if !args[1].(bool) && len(x.ReferencedTweets) != 0 {
				tweets.Tweets[i] = nil
			}
		}
		token = resp.Meta.NextToken
	case 4:
		// Home
		var id string
		id, err = t.getMyId()
		if err != nil {
			break
		}
		excludes := make([]twitter.Exclude, 1, 2)
		excludes[0] = twitter.ExcludeReplies
		if !args[0].(bool) {
			excludes = append(excludes, twitter.ExcludeRetweets)
		}
		var resp *twitter.UserTweetReverseChronologicalTimelineResponse
		resp, err = t.UserTweetReverseChronologicalTimeline(context.Background(), id, twitter.UserTweetReverseChronologicalTimelineOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields:     []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
			MaxResults:      100,
			Excludes:        excludes,
			SinceID:         stopAt,
			PaginationToken: token,
		})
		if err != nil {
			break
		}
		if len(resp.Raw.Errors) != 0 {
			err = errors.New(resp.Raw.Errors[0].Title)
			break
		}
		tweets = resp.Raw
		token = resp.Meta.NextToken
	case 5:
		// Bookmarks
		var id string
		id, err = t.getMyId()
		if err != nil {
			break
		}
		var resp *twitter.TweetBookmarksLookupResponse
		resp, err = t.TweetBookmarksLookup(context.Background(), id, twitter.TweetBookmarksLookupOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields:     []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
			MaxResults:      100,
			PaginationToken: token,
		})
		tweets = resp.Raw
		if stopAt != "" {
			for i, x := range tweets.Tweets {
				if x.ID == stopAt {
					tweets.Tweets = tweets.Tweets[:i]
					break
				}
			}
		}
		token = resp.Meta.NextToken
	default:
		err = errors.New("unknown type")
	}
	if err != nil {
		return nil
	}
	thing.token = token
	nTweets := make([]ImageEntry, 0, len(tweets.Tweets))
	mediaMap := tweets.Includes.MediaByKeys()
	for i, x := range tweets.Tweets {
		if x == nil {
			continue
		}
		var media []*twitter.MediaObj
		if len(x.Attachments.MediaKeys) != 0 {
			media = make([]*twitter.MediaObj, len(x.Attachments.MediaKeys))
			for i2, v := range x.Attachments.MediaKeys {
				media[i2] = mediaMap[v]
			}
		}
		var urls []twitter.EntityURLObj
		if x.Entities != nil {
			urls = x.Entities.URLs
		}
		nTweets = append(nTweets, &TwitterImageEntry{x.ID, x.Text, media, urls, tweets.Includes.Users[i].Name, "", nil})
	}
	return nTweets
}

func (t TwitterSite) GetRequest(URL string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", URL, http.NoBody)
	t.Authorizer.Add(req)
	return t.Client.Client.Do(req)
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
		x2, err := t.TweetLookup(context.Background(), []string{s}, twitter.TweetLookupOpts{
			Expansions:  []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields: []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType},
			TweetFields: []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
		})
		if err != nil {
			return "", nil
		}
		x := x2.Raw.Tweets[0]
		return "", &TwitterImageEntry{x.ID, x.Text, x2.Raw.Includes.Media, x.Entities.URLs, x2.Raw.Includes.Users[0].Name, "", nil}
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
	ID       string
	text     string
	media    []*twitter.MediaObj
	urls     []twitter.EntityURLObj
	username string
	info     string
	data     []ImageEntry
}

func (t *TwitterImageEntry) GetName() string {
	return t.username
}

func (t *TwitterImageEntry) GetPostURL() string {
	return fmt.Sprintf("https://twitter.com/%s/status/%s", t.username, t.ID)
}

func (t *TwitterImageEntry) GetText() string {
	return wordWrapper(t.text)
}

func (t *TwitterImageEntry) GetDimensions() (int, int) {
	if len(t.media) == 0 {
		return -1, -1
	}
	return t.media[0].Width, t.media[0].Height
}

func (t *TwitterImageEntry) GetInfo() string {
	if t.info == "" {
		t.info = fmt.Sprintf("%s\n@%s", t.text, t.username)
	}
	return t.info
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
		s := t.media[0].URL
		ind := strings.LastIndexByte(s, '.')
		return fmt.Sprintf("%s?format=%s&name=large", s[:ind], s[ind+1:])
	} else if len(t.urls) > 0 {
		return t.urls[0].ExpandedURL
	} else {
		return t.GetPostURL()
	}
}

func (t *TwitterImageEntry) Combine(ie ImageEntry) {
	t.GetInfo()
	s := ie.GetInfo()
	if s != "" {
		t.info += "\n" + s
	}
	t.urls = []twitter.EntityURLObj{{ExpandedURL: ie.GetURL()}}
	s = ie.GetName()
	if s != "" {
		t.username = s
	}
	s = ie.GetText()
	if s != "" {
		t.text += "\n" + s
		if ie.GetType() == IETYPE_TEXT {
			t.media = nil
		}
	}
	t.data = ie.GetGalleryInfo(false)
	if len(t.data) == 0 {
		t.data = nil
	}
}

func (t *TwitterImageEntry) GetGalleryInfo(b bool) []ImageEntry {
	if t.data != nil {
		return t.data
	}
	imgs := make([]ImageEntry, len(t.media))
	if b {
		return imgs
	}
	for i, x := range t.media {
		imgs[i] = &TwitterGalleryEntry{*t, x.URL, x.Width, x.Height, i + 1}
	}
	return imgs
}

func (t *TwitterImageEntry) GetSaveName() string {
	if len(t.media) == 0 {
		return ""
	}
	s := t.media[0].URL
	ind := strings.LastIndexByte(s, '/')
	s = s[ind+1:]
	return s
}

type TwitterGalleryEntry struct {
	TwitterImageEntry
	url   string
	w, h  int
	index int
}

func (t *TwitterGalleryEntry) GetURL() string {
	return t.url
}

func (t *TwitterGalleryEntry) GetDimensions() (int, int) {
	return t.w, t.h
}

func (*TwitterGalleryEntry) GetType() ImageEntryType { return IETYPE_REGULAR }

func (t *TwitterGalleryEntry) GetSaveName() string {
	s := t.TwitterImageEntry.GetSaveName()
	ind := strings.LastIndexByte(s, '.')
	return fmt.Sprintf("%s_%d.%s", s[:ind], t.index, s[ind+1:])
}
