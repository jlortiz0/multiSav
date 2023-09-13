package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

type IllustrationType string

const (
	ILTYPE_ILUST  IllustrationType = "illust"
	ILTYPE_UGOIRA IllustrationType = "ugoira"
	ILTYPE_MANGA  IllustrationType = "manga"
	// ILTYPE_NOVEL  IllustrationType = "novel"
	ILTYPE_NONE IllustrationType = ""
)

type Illustration struct {
	Create_date time.Time
	client      *Client
	User        *User
	Image_urls  multisize
	Series      struct {
		Title string
		ID    int
	}
	Title            string
	Type             IllustrationType
	Caption          string
	Meta_single_page struct {
		Original_image_url string
	}
	Tags []struct {
		Name, Translated_name string
	}
	Meta_pages []struct {
		Image_urls multisize
	}
	ID int
	// What does this mean? Is it an R18 bool?
	Restrict   int
	Page_count int
	Width      int
	Height     int
	// What do these mean?
	Sanity_level    int
	X_restrict      int
	Total_view      int
	Total_bookmarks int
	Total_comments  int
	Is_bookmarked   bool
	Is_muted        bool
}

type Comment struct {
	Date    time.Time
	User    *User
	Comment string
	ID      int
}

type UgoiraMetadata struct {
	Zip_urls multisize
	Frames   []struct {
		File  string
		Delay int
	}
}

func (p *Client) IllustFromID(ID int) *Illustration {
	return &Illustration{ID: ID, client: p}
}

func (i *Illustration) Fetch() error {
	if i.Title == "" {
		data, err := i.client.GetIllust(i.ID)
		if err == nil {
			*i = *data
		}
		return err
	}
	return nil
}

type Comments struct {
	Comments       []*Comment
	Total_comments int
	Offset         int
}

func (i *Illustration) GetComments(offset int) (*Comments, error) {
	b := new(strings.Builder)
	b.WriteString("1/illust/comments?illust_id=")
	b.WriteString(strconv.Itoa(i.ID))
	if offset != 0 {
		b.WriteString("&offset=")
		b.WriteString(strconv.Itoa(offset))
	}
	b.WriteString("&include_total_comments=true")
	resp, err := i.client.doGetRequest(b.String())
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output struct {
		Next_url string
		Error    struct {
			Message string
		}
		Comments
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
		output.Comments.Offset = -1
		return &output.Comments, nil
	}
	s := output.Next_url[ind+7:]
	ind = strings.IndexByte(s, '&')
	if ind != -1 {
		s = s[:ind]
	}
	output.Offset, err = strconv.Atoi(s)
	return &output.Comments, err
}

func (i *Illustration) GetRelated() (*IllustrationListing, error) {
	return i.client.newIllustrationListing("2/illust/related?filter=for_ios&illust_id=" + strconv.Itoa(i.ID))
}

func (i *Illustration) Bookmark(visi Visibility) error {
	b := new(strings.Builder)
	b.WriteString("illust_id=")
	b.WriteString(strconv.Itoa(i.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	req := i.client.buildPostRequest("2/illust/bookmark/add", b.String())
	_, err := i.client.client.Do(req)
	return err
}

func (i *Illustration) Unbookmark() error {
	req := i.client.buildPostRequest("1/illust/bookmark/delete", "illust_id="+strconv.Itoa(i.ID))
	_, err := i.client.client.Do(req)
	return err
}

func (i *Illustration) GetUgoiraMetadata() (*UgoiraMetadata, error) {
	resp, err := i.client.doGetRequest("1/ugoira/metadata?illust_id=" + strconv.Itoa(i.ID))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var output struct {
		Error struct {
			Message string
		}
		Ugoira_metadata UgoiraMetadata
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	return &output.Ugoira_metadata, err
}
