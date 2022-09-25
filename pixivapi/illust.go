package pixivapi

import (
	"encoding/json"
	"errors"
	"io"
	"os"
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
	ID         int
	Title      string
	Type       IllustrationType
	Image_urls multisize
	Caption    string
	// What does this mean?
	Restrict int
	User     *User
	// Traslated_name always seems to be defined but null
	Tags []struct {
		Name, Translated_name string
	}
	Create_date time.Time
	Page_count  int
	Width       int
	Height      int
	// What do these mean?
	Sanity_level int
	X_restrict   int
	Series       struct {
		ID    string
		Title string
	}
	Meta_single_page struct {
		Original_image_url string
	}
	Total_view      int
	Total_bookmarks int
	Is_bookmarked   bool
	Is_muted        bool
	Total_comments  int
	client          *Client
}

type Comment struct {
	ID      int
	Comment string
	Date    time.Time
	User    *User
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

func (i *Illustration) Fetch() {
	if i.Title == "" {
		data, err := i.client.GetIllust(i.ID)
		if err == nil {
			*i = *data
		}
	}
}

type Comments struct {
	Total_comments int
	Offset         int
	Comments       []*Comment
}

func (i *Illustration) GetComments(offset int) (*Comments, error) {
	b := new(strings.Builder)
	b.WriteString("illust/comments?illust_id=")
	b.WriteString(strconv.Itoa(i.ID))
	if offset != 0 {
		b.WriteString("&offset=")
		b.WriteString(strconv.Itoa(offset))
	}
	b.WriteString("&include_total_comments=true")
	req := i.client.buildGetRequest(b.String())
	resp, err := i.client.client.Do(req)
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
		Comments
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
	output.Offset, err = strconv.Atoi(s)
	return &output.Comments, err
}

func (i *Illustration) GetRelated() (*IllustrationListing, error) {
	return i.client.newIllustrationListing("illust/related?filter=for_ios&illust_id=" + strconv.Itoa(i.ID))
}

func (i *Illustration) Bookmark(visi Visibility) error {
	b := new(strings.Builder)
	b.WriteString("illust_id=")
	b.WriteString(strconv.Itoa(i.ID))
	b.WriteString("&restrict=")
	b.WriteString(string(visi))
	req := i.client.buildPostRequest("illust/bookmark/add", strings.NewReader(b.String()))
	_, err := i.client.client.Do(req)
	return err
}

func (i *Illustration) Unbookmark() error {
	req := i.client.buildPostRequest("illust/bookmark/delete", strings.NewReader("illust_id="+strconv.Itoa(i.ID)))
	_, err := i.client.client.Do(req)
	return err
}

func (i *Illustration) GetUgoiraMetadata() (*UgoiraMetadata, error) {
	req := i.client.buildGetRequest("ugoira/metadata?illust_id=" + strconv.Itoa(i.ID))
	resp, err := i.client.client.Do(req)
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
		Ugoira_metadata UgoiraMetadata
		Error           struct {
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
	return &output.Ugoira_metadata, err
}
