package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Visibility string

const VISI_PUBLIC Visibility = "public"
const VISI_PRIVATE Visibility = "private"

type multisize struct {
	Square_medium string
	Medium        string
	Large         string
	Original      string
}

func (m multisize) Best() string {
	if m.Original != "" {
		return m.Original
	}
	if m.Large != "" {
		return m.Large
	}
	if m.Medium != "" {
		return m.Medium
	}
	return m.Square_medium
}

type User struct {
	ID                 int
	Name               string
	Account            string
	Profile_image_urls multisize
	Is_followed        bool
	Comment            string
	client             *Client
}

func (p *Client) UserFromID(ID int) *User {
	return &User{ID: ID, client: p}
}

func (u *User) Fetch() error {
	if u.Name == "" {
		data, err := u.client.GetUser(u.ID)
		if err == nil {
			*u = *data
		}
		return err
	}
	return nil
}

func (u *User) Illustrations(kind IllustrationType) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("1/user/illusts?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	if kind != ILTYPE_NONE {
		b.WriteString("&type=")
		b.WriteString(string(kind))
	}
	b.WriteString("&filter=for_ios")
	return u.client.newIllustrationListing(b.String())
}

func (u *User) Bookmarks(tag string, visibility Visibility) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("1/user/bookmarks/illust?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	if tag != "" {
		b.WriteString("&tag=")
		b.WriteString(string(tag))
	}
	b.WriteString("&restrict=")
	b.WriteString(string(visibility))
	b.WriteString("&filter=for_ios")
	return u.client.newIllustrationListing(b.String())
}

type BookmarkTagsResponse struct {
	Bookmark_tags []struct {
		Name  string
		Count int
	}
	Offset int
}

func (u *User) BookmarkTags(visi Visibility, offset int) (*BookmarkTagsResponse, error) {
	b := new(strings.Builder)
	b.WriteString("1/user/bookmark-tags/illust?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	if offset != 0 {
		b.WriteString("&offset=")
		b.WriteString(strconv.Itoa(offset))
	}
	b.WriteString("&filter=for_ios")
	resp, err := u.client.doGetRequest(b.String())
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output struct {
		Bookmark_tags []struct {
			Name  string
			Count int
		}
		Next_url string
		Error    struct {
			Message string
		}
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	ind := strings.Index(output.Next_url, "offset=")
	if ind == -1 {
		return &BookmarkTagsResponse{Bookmark_tags: output.Bookmark_tags}, nil
	}
	s := output.Next_url[ind+7:]
	ind = strings.IndexByte(s, '&')
	if ind != -1 {
		s = s[:ind]
	}
	offset, err = strconv.Atoi(s)
	if err != nil {
		return &BookmarkTagsResponse{Bookmark_tags: output.Bookmark_tags}, err
	}
	return &BookmarkTagsResponse{output.Bookmark_tags, offset}, nil
}

func (u *User) Following(visi Visibility) (*UserListing, error) {
	b := new(strings.Builder)
	b.WriteString("1/user/following?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	return u.client.newUserListing(b.String())
}

// This probably returns empty unless the user is us
func (u *User) Followers() (*UserListing, error) {
	return u.client.newUserListing("1/user/follower?filter=for_ios&user_id=" + strconv.Itoa(u.ID))
}

func (u *User) Related() (*UserListing, error) {
	return u.client.newUserListing("1/user/related?filter=for_ios&seed_user_id=" + strconv.Itoa(u.ID))
}

func (u *User) Follow(visi Visibility) error {
	b := new(strings.Builder)
	b.WriteString("user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	req := u.client.buildPostRequest("1/user/follow/add", b.String())
	_, err := u.client.client.Do(req)
	return err
}

func (u *User) Unfollow() error {
	req := u.client.buildPostRequest("1/user/follow/delete", "user_id="+strconv.Itoa(u.ID))
	_, err := u.client.client.Do(req)
	return err
}
