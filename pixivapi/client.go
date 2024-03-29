package pixivapi

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	cfbp "github.com/DaRealFreak/cloudflare-bp-go"
	"golang.org/x/oauth2"
)

// https://github.com/upbit/pixivpy used as reference for this package

const (
	login_secret  = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"
	auth_url      = "https://oauth.secure.pixiv.net/auth/token"
	base_url      = "https://app-api.pixiv.net/v"
	ios_version   = "15.7"
	user_agent    = "PixivIOSApp/7.15.14 (iOS " + ios_version + "; iPhone13,2)"
	Client_ID     = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	Client_secret = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
)

type RankingMode string

const (
	DAY           RankingMode = "day"
	WEEK          RankingMode = "week"
	MONTH         RankingMode = "month"
	DAY_MALE      RankingMode = "day_male"
	DAY_FEMALE    RankingMode = "day_female"
	WEEK_ORIGINAL RankingMode = "week_original"
	WEEK_ROOKIE   RankingMode = "week_rookie"
	// DAY_MANGA     RankingMode = "day_manga"
)

type SearchTarget string
type SearchSort string
type SearchDuration string

const (
	TAGS_PARTIAL      SearchTarget   = "partial_match_for_tags"
	TAGS_EXACT        SearchTarget   = "exact_match_for_tags"
	TITLE_AND_CAPTION SearchTarget   = "title_and_caption"
	DATE_DESC         SearchSort     = "date_desc"
	DATE_ASC          SearchSort     = "date_asc"
	WITHIN_DAY        SearchDuration = "within_last_day"
	WITHIN_WEEK       SearchDuration = "within_last_week"
	WITHIN_MONTH      SearchDuration = "within_last_month"
	WITHIN_NONE       SearchDuration = ""
)

type Client struct {
	client *http.Client
	token  *oauth2.Token
	myId   int
}

func NewClient() *Client {
	c := new(Client)
	c.client = new(http.Client)
	c.client.Transport = cfbp.AddCloudFlareByPass(nil)
	return c
}

func (p *Client) SetToken(access, refresh string) {
	p.token = &oauth2.Token{AccessToken: access, RefreshToken: refresh, Expiry: time.Now().Add(time.Minute * 10)}
}

func (p *Client) IsLoggedIn() bool {
	return p.token != nil
}

func (p *Client) Login(token string) error {
	ts := time.Now().UTC().Format("2006-01-02T15:04:05") + "+00:00"
	buf := new(bytes.Buffer)
	buf.WriteString("client_id=")
	buf.WriteString(Client_ID)
	buf.WriteString("&client_secret=")
	buf.WriteString(Client_secret)
	buf.WriteString("&grant_type=refresh_token&include_policy=true&refresh_token=")
	buf.WriteString(token)
	req, _ := http.NewRequest("POST", auth_url, buf)
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("app-os", "ios")
	req.Header.Add("app-os-version", ios_version)
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(buf.Len()))
	req.Header.Add("X-Client-Time", ts)
	sum := md5.Sum([]byte(ts + login_secret))
	req.Header.Add("X-Client-Hash", hex.EncodeToString(sum[:]))
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	var output struct {
		User struct {
			ID string
		}
		Access_token      string
		Refresh_token     string
		Error_description string
		Expires_in        int
	}
	mData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(mData, &output)
	if err != nil {
		return err
	}
	if output.Error_description != "" {
		return errors.New(output.Error_description)
	}
	p.myId, err = strconv.Atoi(output.User.ID)
	p.token = &oauth2.Token{AccessToken: output.Access_token, RefreshToken: output.Refresh_token, Expiry: time.Now().Add(time.Second * time.Duration(output.Expires_in))}
	return err
}

func (p *Client) checkToken() error {
	if !p.token.Valid() {
		return p.Login(p.token.RefreshToken)
	}
	return nil
}

func (p *Client) RefreshToken() string {
	if p.token == nil {
		return ""
	}
	return p.token.RefreshToken
}

func (p *Client) doGetRequest(url string) (*http.Response, error) {
	return p.GetRequest(base_url + url)
}

func (p *Client) GetRequest(url string) (*http.Response, error) {
	if p.token == nil {
		return nil, errors.New("not logged in")
	}
	p.checkToken()
	req, _ := http.NewRequest("GET", url, http.NoBody)
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("app-os", "ios")
	req.Header.Add("app-os-version", ios_version)
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Referer", "https://app-api.pixiv.net/")
	p.token.SetAuthHeader(req)
	return p.client.Do(req)
}

func (p *Client) buildPostRequest(url string, body string) *http.Request {
	if p.token == nil {
		return nil
	}
	p.checkToken()
	req, _ := http.NewRequest("POST", base_url+url, strings.NewReader(body))
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("app-os", "ios")
	req.Header.Add("app-os-version", ios_version)
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	p.token.SetAuthHeader(req)
	return req
}

func (p *Client) GetIllust(id int) (*Illustration, error) {
	resp, err := p.doGetRequest("1/illust/detail?illust_id=" + strconv.Itoa(id))
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
		Illust Illustration
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	output.Illust.client = p
	output.Illust.User.client = p
	return &output.Illust, nil
}

func (p *Client) GetUser(id int) (*User, error) {
	resp, err := p.doGetRequest("1/user/detail?filter=for_ios&user_id=" + strconv.Itoa(id))
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
		User User
		// Profile, Profile_publicity, Workspace
	}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return nil, err
	}
	if output.Error.Message != "" {
		return nil, errors.New(output.Error.Message)
	}
	output.User.client = p
	return &output.User, nil
}

func (p *Client) RecommendedIllust(kind IllustrationType) (*IllustrationListing, error) {
	return p.newIllustrationListing("1/illust/recommended?content_type=" + string(kind))
}

func (p *Client) RankedIllust(mode RankingMode, day time.Time) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("1/illust/ranking?mode=")
	b.WriteString(string(mode))
	if !day.IsZero() {
		b.WriteString("&date=")
		b.WriteString(day.Format("2006-01-02"))
	}
	b.WriteString("&filter=for_ios")
	return p.newIllustrationListing(b.String())
}

// func (p *Client) TrendingTags() {}

func (p *Client) SearchIllust(term string, target SearchTarget, sorting SearchSort, duration SearchDuration) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("1/search/illust?word=")
	b.WriteString(term)
	b.WriteString("&search_target=")
	b.WriteString(string(target))
	b.WriteString("&sort=")
	b.WriteString(string(sorting))
	b.WriteString("&filter=for_ios")
	if duration != WITHIN_NONE {
		b.WriteString("&duration=")
		b.WriteString(string(duration))
	}
	return p.newIllustrationListing(b.String())
}

func (p *Client) SearchUser(term string, sorting SearchSort, duration SearchDuration) (*UserListing, error) {
	// How do sort and duration affect this? Most recent posting? Join date?
	b := new(strings.Builder)
	b.WriteString("1/search/user?word=")
	b.WriteString(term)
	b.WriteString("&sort=")
	b.WriteString(string(sorting))
	b.WriteString("&filter=for_ios")
	if duration != WITHIN_NONE {
		b.WriteString("&duration=")
		b.WriteString(string(duration))
	}
	return p.newUserListing(b.String())
}

func (p *Client) GetMyId() int {
	return p.myId
}

func (p *Client) Followed(visi Visibility) (*IllustrationListing, error) {
	return p.newIllustrationListing("2/illust/follow?restrict=" + string(visi))
}
