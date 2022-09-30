package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"jlortiz.org/redisav/pixivapi"
)

type PixivSite struct {
	*pixivapi.Client
}

func NewPixivSite(refresh string) (PixivSite, error) {
	ret := pixivapi.NewClient()
	return PixivSite{ret}, ret.Login(refresh)
}

func (p PixivSite) Destroy() {}

func (p PixivSite) GetListingInfo() []ListingInfo {
	return []ListingInfo{
		{"User", []ListingArgument{{name: "ID or URL"}}},
		{"Bookmarks", []ListingArgument{{name: "User ID or URL (0 for self)"}}},
		{"Search", []ListingArgument{{name: "Query"}, {name: "Search kind", options: []interface{}{
			pixivapi.TAGS_EXACT, pixivapi.TAGS_EXACT, pixivapi.TITLE_AND_CAPTION,
		}}}},
		{"Recommended", nil},
		{"Best of", []ListingArgument{{name: "Of", options: []interface{}{
			pixivapi.DAY, pixivapi.WEEK, pixivapi.MONTH, pixivapi.DAY_MALE, pixivapi.DAY_FEMALE,
		}}}},
	}
}

func (p PixivSite) GetResolvableDomains() []string {
	return []string{"pixiv.net", "i.pximg.net"}
}

type PixivImageListing struct {
	*pixivapi.IllustrationListing
	persist int
	kind    int
	args    []interface{}
}

func (p *PixivImageListing) GetInfo() (int, []interface{}) {
	return p.kind, p.args
}

func (p *PixivImageListing) GetPersistent() interface{} {
	return p.persist
}

func (p PixivSite) GetListing(kind int, args []interface{}, persist interface{}) (ImageListing, []ImageEntry) {
	var ls *pixivapi.IllustrationListing
	var err error
	switch kind {
	case 0:
		// User
		fallthrough
	case 1:
		// Bookmarks
		s := args[0].(string)
		ind := strings.Index(s, "users/")
		if ind != -1 {
			s = s[ind+6:]
			ind = strings.IndexByte(s, '?')
			if ind != -1 {
				s = s[:ind]
			}
		}
		var i int
		i, err = strconv.Atoi(s)
		if err == nil {
			if i == 0 {
				i = p.GetMyId()
			}
			u := p.UserFromID(i)
			if kind == 0 {
				ls, err = u.Illustrations(pixivapi.ILTYPE_ILUST)
			} else {
				ls, err = u.Bookmarks("", pixivapi.VISI_PRIVATE)
			}
		}
	case 2:
		// Search
		ls, err = p.SearchIllust(args[0].(string), args[1].(pixivapi.SearchTarget), pixivapi.DATE_DESC, pixivapi.WITHIN_NONE)
	case 3:
		// Recommended
		ls, err = p.RecommendedIllust(pixivapi.ILTYPE_ILUST)
	case 4:
		// Best of
		ls, err = p.RankedIllust(args[0].(pixivapi.RankingMode), time.Time{})
	default:
		err = errors.New("unknown kind")
	}
	if err != nil {
		return ErrorListing{err}, nil
	}
	out := &PixivImageListing{ls, kind, 0, args}
	if persist != nil {
		out.persist = int(persist.(float64))
	}
	return out, p.ExtendListing(out)
}

func (p PixivSite) ExtendListing(ls ImageListing) []ImageEntry {
	iter2, ok := ls.(*PixivImageListing)
	if !ok {
		return nil
	}
	iter := iter2.IllustrationListing
	if iter == nil {
		return nil
	}
	x, err := iter.Next()
	if err != nil || x == nil {
		return nil
	}
	data := make([]ImageEntry, 1, iter.Buffered()+1)
	data[0] = &PixivImageEntry{Illustration: x}
	for !iter.NextRequiresFetch() {
		x, err = iter.Next()
		if err == nil {
			if x == nil {
				break
			}
			data = append(data, &PixivImageEntry{Illustration: x})
			if x.ID == iter2.persist {
				iter2.IllustrationListing = nil
				break
			}
		}
	}
	return data
}

func (p PixivSite) ResolveURL(URL string) (string, ImageEntry) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", nil
	}
	switch u.Hostname() {
	case "www.pixiv.net":
		fallthrough
	case "pixiv.net":
		s := u.Path
		ind := strings.Index(s, "artworks")
		if ind == -1 {
			return "", nil
		}
		s = s[ind+9:]
		i, err := strconv.Atoi(s)
		if err != nil {
			return "", nil
		}
		x2, err := p.GetIllust(i)
		if err != nil {
			return "", nil
		}
		return "", &PixivImageEntry{Illustration: x2}
	case "i.pximg.net":
		fallthrough
	case "pximg.net":
		return RESOLVE_FINAL, nil
	}
	return "", nil
}

type PixivImageEntry struct {
	*pixivapi.Illustration
}

func (p *PixivImageEntry) GetName() string {
	return p.Title
}

func (p *PixivImageEntry) GetURL() string {
	if p.Meta_single_page.Original_image_url != "" {
		return p.Meta_single_page.Original_image_url
	}
	if p.Image_urls.Original != "" {
		return p.Image_urls.Original
	}
	if p.Image_urls.Large != "" {
		return p.Image_urls.Large
	}
	return p.Image_urls.Medium
}

func (p *PixivImageEntry) GetGalleryInfo(b bool) []ImageEntry {
	arr := make([]ImageEntry, p.Page_count)
	if b {
		return arr
	}
	for i := range p.Meta_pages {
		arr[i] = &PixivGalleryEntry{*p, i}
	}
	return arr
}

func (p *PixivImageEntry) GetType() ImageEntryType {
	if p.Page_count > 1 {
		return IETYPE_GALLERY
	}
	// if p.type == pixivapi.ILTYPE_NOVEL
	return IETYPE_REGULAR
}

func (p *PixivImageEntry) GetDimensions() (int, int) {
	return p.Width, p.Height
}

func (p *PixivImageEntry) GetPostURL() string {
	return "https://pixiv.net/en/artworks/" + strconv.Itoa(p.ID)
}

func (p *PixivImageEntry) GetInfo() string {
	return fmt.Sprintf("%s by %s\n%s\nView: %d  Bookmark: %d  Comments: %d", p.Title, p.User.Name, p.Caption, p.Total_view, p.Total_bookmarks, p.Total_comments)
}

func (p *PixivImageEntry) GetSaveName() string {
	s := p.GetURL()
	ind := strings.LastIndexByte(s, '/')
	return s[ind+1:]
}

func (*PixivImageEntry) GetText() string { return "" }

func (p *PixivImageEntry) Combine(ImageEntry) {
	panic("this should never occur")
}

type PixivGalleryEntry struct {
	PixivImageEntry
	ind int
}

func (p *PixivGalleryEntry) GetURL() string {
	if p.Meta_pages[p.ind].Image_urls.Original != "" {
		return p.Meta_pages[p.ind].Image_urls.Original
	}
	if p.Meta_pages[p.ind].Image_urls.Large != "" {
		return p.Meta_pages[p.ind].Image_urls.Large
	}
	return p.Meta_pages[p.ind].Image_urls.Medium
}

func (p *PixivGalleryEntry) GetSaveName() string {
	return p.GetURL()
}
