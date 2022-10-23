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
	// What does this mean? Is it an R18 bool?
	Restrict int
	User     *User
	Tags     []struct {
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
		ID    int
		Title string
	}
	Meta_single_page struct {
		Original_image_url string
	}
	Meta_pages []struct {
		Image_urls multisize
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
	Total_comments int
	Offset         int
	Comments       []*Comment
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
	resp, err := i.client.client.Do(req)
	if debug_output {
		f, _ := os.OpenFile("out.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		f.WriteString(b.String())
		f.Write([]byte{'\n'})
		f.WriteString(resp.Request.URL.Path)
		f.Write([]byte{'\n'})
		io.Copy(f, resp.Body)
		f.Close()
	}
	return err
}

func (i *Illustration) Unbookmark() error {
	req := i.client.buildPostRequest("1/illust/bookmark/delete", "illust_id="+strconv.Itoa(i.ID))
	resp, err := i.client.client.Do(req)
	if debug_output {
		f, _ := os.OpenFile("out.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		f.WriteString("illust_id=")
		f.WriteString(strconv.Itoa(i.ID))
		f.Write([]byte{'\n'})
		io.Copy(f, resp.Body)
		f.Close()
	}
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
