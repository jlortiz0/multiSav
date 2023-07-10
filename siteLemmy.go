package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"go.elara.ws/go-lemmy"
	"go.elara.ws/go-lemmy/types"
)

type LemmySite struct {
	*lemmy.Client
	base string
}

func NewLemmyClient(site string, user string, pass string) LemmySite {
	if site == "" {
		return LemmySite{}
	}
	cl, err := lemmy.New("https://" + site)
	if err != nil {
		panic(err)
	}
	if user != "" {
		err = cl.ClientLogin(context.Background(), types.Login{UsernameOrEmail: user, Password: pass})
		if err != nil {
			panic(err)
		}
	}
	return LemmySite{cl, site}
}

func (LemmySite) Destroy() {}

func (l LemmySite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{
			name: "New: community",
			args: []ListingArgument{
				{
					name: "Community",
				},
			},
		},
	}
}

type LemmyPostListing struct {
	types.GetPosts
	kind       int
	args       []interface{}
	persistent int
	page       int64
}

func (p *LemmyPostListing) GetInfo() (int, []interface{}) {
	return p.kind, p.args
}

func (p *LemmyPostListing) GetPersistent() interface{} {
	return float64(p.persistent)
}

func (lem LemmySite) GetListing(kind int, args []interface{}, persist interface{}) (ImageListing, []ImageEntry) {
	if lem.Client == nil {
		return ErrorListing{errors.New("not signed in to Lemmy")}, nil
	}
	var posts types.GetPosts
	var err error
	switch kind {
	case 0:
		var sub *types.GetCommunityResponse
		sub, err = lem.Community(context.Background(), types.GetCommunity{Name: types.NewOptional[string](args[0].(string))})
		if err == nil && sub.Error.IsValid() {
			err = errors.New(sub.Error.String())
		}
		if err == nil {
			posts = types.GetPosts{
				Sort:        types.NewOptional[types.SortType](types.SortTypeNew),
				CommunityID: types.NewOptional[int](sub.CommunityView.Counts.CommunityID),
			}
		}
	}
	if err != nil {
		return ErrorListing{err}, nil
	}
	posts.Limit = types.NewOptional[int64](20)
	ls := &LemmyPostListing{GetPosts: posts, kind: kind, args: args}
	if persist != nil {
		temp := persist.(float64)
		ls.persistent = int(temp)
	}
	out, err := lem.ExtendListing(ls)
	if err != nil {
		return ErrorListing{err}, nil
	}
	return ls, out
}

func (lem LemmySite) ExtendListing(cont ImageListing) ([]ImageEntry, error) {
	posts, ok := cont.(*LemmyPostListing)
	if !ok {
		return nil, errors.New("unable to cast listing")
	}
	if posts.page == -1 {
		return nil, nil
	}
	posts.page += 1
	posts.Page = types.NewOptional[int64](posts.page)
	resp, err := lem.Posts(context.Background(), posts.GetPosts)
	if err != nil {
		return nil, err
	}
	if resp.Error.IsValid() {
		return nil, errors.New(resp.Error.String())
	}
	ls := make([]ImageEntry, 0, len(resp.Posts))
	for _, v := range resp.Posts {
		if v.Post.ID < posts.persistent {
			posts.page = -1
			break
		}
		ls = append(ls, LemmyImageEntry{v.Post, lem.base})
	}
	return ls, nil
}

func (lem LemmySite) GetRequest(url string) (*http.Response, error) {
	return http.Get(url)
}

func (lem LemmySite) GetResolvableDomains() []string {
	// TODO: ???????????????????
	// The problem with the fediverse...
	return []string{"feddit.nl", "burggit.moe"}
}

func (lem LemmySite) ResolveURL(u string) (string, ImageEntry) {
	ind := strings.LastIndexByte(u, '/')
	if ind == -1 {
		return "", nil
	}
	ind2 := strings.LastIndexByte(u[:ind], '/')
	if u[ind2+1:ind] != "post" {
		return "", nil
	}
	id, err := strconv.Atoi(u[ind+1:])
	if err != nil {
		return "", nil
	}
	ps, err := lem.Post(context.Background(), types.GetPost{
		ID: types.NewOptional[int](id),
	})
	if err != nil || ps.Error.IsValid() {
		return "", nil
	}
	return "", LemmyImageEntry{ps.PostView.Post, lem.base}
}

type LemmyImageEntry struct {
	types.Post
	site string
}

func (l LemmyImageEntry) GetName() string { return l.Name }

func (l LemmyImageEntry) GetText() string { return l.Body.ValueOr("") }

func (l LemmyImageEntry) GetURL() string { return l.URL.ValueOr("") }

func (l LemmyImageEntry) GetGalleryInfo(bool) []ImageEntry { return nil }

func (l LemmyImageEntry) GetType() ImageEntryType {
	if l.URL.IsValid() {
		return IETYPE_REGULAR
	}
	return IETYPE_TEXT
}

func (l LemmyImageEntry) GetPostURL() string {
	return "https://" + l.site + "/post/" + strconv.Itoa(l.ID)
}

func (l LemmyImageEntry) GetSaveName() string {
	if !l.URL.IsValid() {
		return ""
	}
	u := l.URL.MustValue()
	ind := strings.LastIndexByte(u, '/')
	if ind == -1 {
		return ""
	}
	s := u[ind+1:]
	if s == "" {
		return s
	}
	ind = strings.IndexByte(s, '?')
	if ind != -1 {
		s = s[:ind]
	}
	return s
}

// TODO
func (l LemmyImageEntry) GetInfo() string { return "" }

// TODO: Support imgur maybe
func (l LemmyImageEntry) Combine(ie ImageEntry) {
	if ie.GetText() != "" {
		if l.Body.IsValid() {
			l.Body.Set(l.Body.MustValue() + "\n" + ie.GetText())
		} else {
			l.Body.Set(ie.GetText())
		}
	}
	l.URL.Set(ie.GetURL())
}

type LemmyProducer struct {
	*BufferedImageProducer
	site LemmySite
}

func NewLemmyProducer(site LemmySite, kind int, args []interface{}, persistent interface{}) LemmyProducer {
	return LemmyProducer{NewBufferedImageProducer(site, kind, args, persistent), site}
}

func (p LemmyProducer) ActionHandler(key int32, sel int, call int) ActionRet {
	useful, ok := p.items[sel].(LemmyImageEntry)
	if !ok {
		return p.BufferedImageProducer.ActionHandler(key, sel, call)
	}
	switch key {
	case rl.KeyL:
		if p.listing.(*LemmyPostListing).kind != 1 {
			p.listing.(*LemmyPostListing).persistent = useful.ID
			p.items = p.items[:sel+1]
			for i := BIP_BUFBEFORE + 1; i < BIP_BUFBEFORE+1+BIP_BUFAFTER; i++ {
				p.buffer[i] = BufferObject{}
			}
			p.listing.(*LemmyPostListing).page = -1
			return ARET_MOVEUP
		}
	case rl.KeyEnter:
		ret := p.BufferedImageProducer.ActionHandler(key, sel, call)
		rl.SetWindowTitle(p.GetTitle())
		return ret
	}
	return p.BufferedImageProducer.ActionHandler(key, sel, call)
}

func (p LemmyProducer) GetTitle() string {
	if p.listing == nil {
		return p.BufferedImageProducer.GetTitle()
	}
	return "multiSav - Lemmy"
}
