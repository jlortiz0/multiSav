package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/oauth2"
)

type TwitterSite struct {
	*twitter.Client
}

type twitterAuthorizer struct {
	b bool
}

func (t twitterAuthorizer) Add(req *http.Request) {}

func NewTwitterSite(refresh string) TwitterSite {
	if refresh == "" {
		return TwitterSite{&twitter.Client{Authorizer: twitterAuthorizer{false}, Client: oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: TwitterBearer,
		})), Host: "https://api.twitter.com"}}
	}
	config := &oauth2.Config{
		ClientID:     TwitterID,
		ClientSecret: TwitterSecret,
		RedirectURL:  "http://localhost:5738/twitter",
		Scopes:       []string{"tweet.read", "users.read", "list.read", "offline.access", "bookmark.read", "bookmark.write"},
	}
	config.Endpoint.AuthURL = "https://twitter.com/i/oauth2/authorize"
	config.Endpoint.TokenURL = "https://api.twitter.com/2/oauth2/token"
	token := &oauth2.Token{RefreshToken: refresh}
	return TwitterSite{&twitter.Client{Authorizer: twitterAuthorizer{true}, Client: config.Client(context.Background(), token), Host: "https://api.twitter.com"}}
}

func (t TwitterSite) Destroy() {
	s, ok := t.Client.Client.Transport.(*oauth2.Transport)
	if ok {
		s2, _ := s.Source.Token()
		if s2 != nil {
			saveData.Twitter = s2.RefreshToken
		}
	}
}

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
	myId    string
}

func (t *TwitterImageListing) GetInfo() (int, []interface{}) {
	return t.kind, t.args
}

func (t *TwitterImageListing) GetPersistent() interface{} {
	return t.persist
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
	me, err := t.AuthUserLookup(context.Background(), twitter.UserLookupOpts{})
	if err == nil {
		thing.myId = me.Raw.Users[0].ID
	}
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
		id := args[0].(string)
		_, err = strconv.Atoi(id)
		if err != nil {
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
			id = user.Raw.Users[0].ID
			args[0] = id
		}
		excludes := make([]twitter.Exclude, 1, 2)
		excludes[0] = twitter.ExcludeReplies
		if !args[1].(bool) {
			excludes = append(excludes, twitter.ExcludeRetweets)
		}
		var resp *twitter.UserTweetTimelineResponse
		resp, err = t.UserTweetTimeline(context.Background(), id, twitter.UserTweetTimelineOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
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
			MediaFields: []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
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
		ind := strings.Index(s, "lists/")
		if ind != -1 {
			s = s[ind+6:]
			args[0] = s
		}
		var resp *twitter.ListTweetLookupResponse
		resp, err = t.ListTweetLookup(context.Background(), s, twitter.ListTweetLookupOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
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
		excludes := make([]twitter.Exclude, 1, 2)
		excludes[0] = twitter.ExcludeReplies
		if !args[0].(bool) {
			excludes = append(excludes, twitter.ExcludeRetweets)
		}
		var resp *twitter.UserTweetReverseChronologicalTimelineResponse
		resp, err = t.UserTweetReverseChronologicalTimeline(context.Background(), thing.myId, twitter.UserTweetReverseChronologicalTimelineOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
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
		var resp *twitter.TweetBookmarksLookupResponse
		resp, err = t.TweetBookmarksLookup(context.Background(), thing.myId, twitter.TweetBookmarksLookupOpts{
			Expansions:      []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields:     []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
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
	userMap := tweets.Includes.UsersByID()
	for _, x := range tweets.Tweets {
		if x == nil {
			continue
		}
		var media []*twitter.MediaObj
		if x.Attachments != nil && len(x.Attachments.MediaKeys) != 0 {
			media = make([]*twitter.MediaObj, len(x.Attachments.MediaKeys))
			for i2, v := range x.Attachments.MediaKeys {
				media[i2] = mediaMap[v]
			}
		}
		var urls []twitter.EntityURLObj
		if x.Entities != nil {
			urls = x.Entities.URLs
		}
		nTweets = append(nTweets, &TwitterImageEntry{x.ID, x.Text, media, urls, userMap[x.AuthorID].UserName, "", nil})
	}
	return nTweets
}

func (t TwitterSite) GetRequest(URL string) (*http.Response, error) {
	URL = strings.ReplaceAll(URL, "&amp;", "&")
	req, _ := http.NewRequest("GET", URL, http.NoBody)
	if !strings.Contains(URL, "pbs.twimg") {
		t.Authorizer.Add(req)
	}
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
	case "mobile.twitter.com":
		fallthrough
	case "twitter.com":
		s := u.Path
		ind := strings.Index(s, "status")
		if ind == -1 {
			return "", nil
		}
		s = s[ind+7:]
		ind = strings.IndexByte(s, '/')
		if ind != -1 {
			s = s[:ind]
		}
		x2, err := t.TweetLookup(context.Background(), []string{s}, twitter.TweetLookupOpts{
			Expansions:  []twitter.Expansion{twitter.ExpansionAttachmentsMediaKeys, twitter.ExpansionAuthorID},
			MediaFields: []twitter.MediaField{twitter.MediaFieldHeight, twitter.MediaFieldMediaKey, twitter.MediaFieldURL, twitter.MediaFieldWidth, twitter.MediaFieldType, twitter.MediaFieldVariants},
			TweetFields: []twitter.TweetField{twitter.TweetFieldAuthorID, twitter.TweetFieldAttachments, twitter.TweetFieldEntities, twitter.TweetFieldID, twitter.TweetFieldText},
		})
		if err != nil || len(x2.Raw.Errors) != 0 {
			return "", nil
		}
		x := x2.Raw.Tweets[0]
		return "", &TwitterImageEntry{x.ID, x.Text, x2.Raw.Includes.Media, x.Entities.URLs, x2.Raw.Includes.Users[0].Name, "", nil}
	case "pbs.twimg.com":
		fallthrough
	case "video.twimg.com":
		fallthrough
	case "twimg.com":
		return RESOLVE_FINAL, nil
	}
	return "", nil
}

func (t TwitterSite) GetResolvableDomains() []string {
	return []string{"twitter.com", "pbs.twimg.com", "twimg.com", "www.twitter.com", "mobile.twitter.com", "video.twimg.com"}
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

func (t *TwitterImageEntry) GetInfo() string {
	if t.info == "" {
		t.info = fmt.Sprintf("%s\n@%s", t.text, t.username)
	}
	return t.info
}

func (t *TwitterImageEntry) GetType() ImageEntryType {
	if len(t.media) > 2 {
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
		if s == "" {
			return t.media[0].Variants[0].URL
		}
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
		imgs[i] = &TwitterGalleryEntry{*t, x.URL, i + 1}
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
	index int
}

func (t *TwitterGalleryEntry) GetURL() string {
	return t.url
}

func (*TwitterGalleryEntry) GetType() ImageEntryType { return IETYPE_REGULAR }

func (t *TwitterGalleryEntry) GetSaveName() string {
	s := t.TwitterImageEntry.GetSaveName()
	ind := strings.LastIndexByte(s, '.')
	return fmt.Sprintf("%s_%d.%s", s[:ind], t.index, s[ind+1:])
}

func (*TwitterGalleryEntry) GetGalleryInfo(bool) []ImageEntry { return nil }

type TwitterProducer struct {
	*BufferedImageProducer
	site TwitterSite
}

func NewTwitterProducer(site TwitterSite, kind int, args []interface{}, persist interface{}) TwitterProducer {
	return TwitterProducer{NewBufferedImageProducer(site, kind, args, persist), site}
}

func (t TwitterProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	var useful *TwitterImageEntry
	switch v := t.items[sel].(type) {
	case *TwitterImageEntry:
		useful = v
	case *TwitterGalleryEntry:
		useful = &v.TwitterImageEntry
	default:
		return t.BufferedImageProducer.ActionHandler(key, sel, call)
	}
	if key == rl.KeyX {
		if !t.site.Authorizer.(twitterAuthorizer).b || saveData.Settings.SaveOnX || t.listing.(*TwitterImageListing).kind == 5 || rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || t.listing.(*TwitterImageListing).myId == "" {
			ret := t.BufferedImageProducer.ActionHandler(key, sel, call)
			if ret&ARET_REMOVE != 0 {
				t.site.RemoveTweetBookmark(context.Background(), t.listing.(*TwitterImageListing).myId, useful.ID)
			}
			return ret
		} else {
			t.site.AddTweetBookmark(context.Background(), t.listing.(*TwitterImageListing).myId, useful.ID)
		}
		t.remove(sel)
		return ARET_MOVEUP | ARET_REMOVE
	} else if key == rl.KeyC {
		if t.listing.(*TwitterImageListing).kind == 5 {
			t.site.RemoveTweetBookmark(context.Background(), t.listing.(*TwitterImageListing).myId, useful.ID)
		}
		t.remove(sel)
		return ARET_MOVEDOWN | ARET_REMOVE
	} else if key == rl.KeyL {
		if t.listing.(*RedditImageListing).kind != 5 {
			t.listing.(*TwitterImageListing).persist = useful.ID
			t.items = t.items[:sel+1]
			for i := BIP_BUFBEFORE + 1; i < BIP_BUFBEFORE+1+BIP_BUFAFTER; i++ {
				t.buffer[i] = nil
			}
			return ARET_MOVEUP
		}
	} else if key == rl.KeyEnter {
		ret := t.BufferedImageProducer.ActionHandler(key, sel, call)
		rl.SetWindowTitle(t.GetTitle())
		return ret
	}
	return t.BufferedImageProducer.ActionHandler(key, sel, call)
}

func (t TwitterProducer) GetTitle() string {
	if t.listing == nil {
		return t.BufferedImageProducer.GetTitle()
	}
	i, args := t.listing.GetInfo()
	switch i {
	case 0:
		u, err := t.site.UserLookup(context.Background(), []string{args[0].(string)}, twitter.UserLookupOpts{})
		if err != nil {
			return err.Error()
		}
		if len(u.Raw.Errors) != 0 {
			return u.Raw.Errors[0].Title
		}
		return "multiSav - Twitter - Timeline: @" + u.Raw.Users[0].Name
	case 1:
		return "multiSav - Twitter - Search: " + args[0].(string)
	case 3:
		l, err := t.site.ListLookup(context.Background(), args[0].(string), twitter.ListLookupOpts{Expansions: []twitter.Expansion{twitter.ExpansionOwnerID}})
		if err != nil {
			return err.Error()
		}
		if len(l.Raw.Errors) != 0 {
			return l.Raw.Errors[0].Title
		}
		return "multiSav - Twitter - List: @" + l.Raw.Includes.Users[0].Name + "/" + l.Raw.List.Name
	case 4:
		u, err := t.site.AuthUserLookup(context.Background(), twitter.UserLookupOpts{})
		if err != nil {
			return err.Error()
		}
		if len(u.Raw.Errors) != 0 {
			return u.Raw.Errors[0].Title
		}
		return "multiSav - Twitter - Home: @" + u.Raw.Users[0].Name
	case 5:
		u, err := t.site.AuthUserLookup(context.Background(), twitter.UserLookupOpts{})
		if err != nil {
			return err.Error()
		}
		if len(u.Raw.Errors) != 0 {
			return u.Raw.Errors[0].Title
		}
		return "multiSav - Twitter - Bookmarks: @" + u.Raw.Users[0].Name
	default:
		return "multiSav - Twitter - Unknown"
	}
}
