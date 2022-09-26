package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"os"
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

type Account struct {
	User
	Mail_address string
	Is_premium   bool
	// TODO: What do these mean?
	X_restrict         int
	Is_mail_authorized bool
}

func (p *Client) UserFromID(ID int) *User {
	return &User{ID: ID, client: p}
}

func (u *User) Fetch() {
	if u.Name == "" {
		data, err := u.client.GetUser(u.ID)
		if err == nil {
			*u = *data
		}
	}
}

func (u *User) Illustrations(kind IllustrationType) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("user/illusts?user_id=")
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
	b.WriteString("user/bookmarks/illust?user_id=")
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
	b.WriteString("user/bookmark-tags/illust?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	if offset != 0 {
		b.WriteString("&offset=")
		b.WriteString(strconv.Itoa(offset))
	}
	b.WriteString("&filter=for_ios")
	req := u.client.buildGetRequest(b.String())
	resp, err := u.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if debug_output {
		f, _ := os.OpenFile("out.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		f.Write(data)
		f.Close()
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
	b.WriteString("user/following?user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	return u.client.newUserListing(b.String())
}

func (u *User) Followers(visi Visibility) (*UserListing, error) {
	return u.client.newUserListing("user/following?filter=for_ios&user_id=" + strconv.Itoa(u.ID))
}

func (u *User) Related() (*UserListing, error) {
	return u.client.newUserListing("user/related?filter=for_ios&seed_user_id=" + strconv.Itoa(u.ID))
}

func (u *User) Follow(visi Visibility) error {
	b := new(strings.Builder)
	b.WriteString("user_id=")
	b.WriteString(strconv.Itoa(u.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	req := u.client.buildPostRequest("user/follow/add", strings.NewReader(b.String()))
	_, err := u.client.client.Do(req)
	return err
}

func (u *User) Unfollow() error {
	req := u.client.buildPostRequest("user/follow/delete", strings.NewReader("user_id="+strconv.Itoa(u.ID)))
	_, err := u.client.client.Do(req)
	return err
}

// What does this retur?
func (u *User) MyPixiv() (*UserListing, error) {
	return u.client.newUserListing("user/mypixiv?user_id=" + strconv.Itoa(u.ID))
}