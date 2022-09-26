package pixivapi

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cfbp "github.com/DaRealFreak/cloudflare-bp-go"
)

const login_secret = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"
const auth_url = "https://oauth.secure.pixiv.net/auth/token"
const base_url = "https://app-api.pixiv.net/v1/"
const ios_version = "15.7"
const user_agent = "PixivIOSApp/7.15.11 (iOS " + ios_version + "; iPhone13,2)"
const debug_output = true

type RankingMode string

const (
	DAY           RankingMode = "day"
	WEEK          RankingMode = "week"
	MONTH         RankingMode = "month"
	DAY_MALE      RankingMode = "day_male"
	DAY_FEMALE    RankingMode = "day_female"
	WEEK_ORIGINAL RankingMode = "week_original"
	WEEK_ROOKIE   RankingMode = "week_rookie"
	DAY_MANGA     RankingMode = "day_manga"
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
	clientId     string
	clientSecret string
	accessToken  string
	refreshToken string
	client       *http.Client
}

type authRequest struct {
	Client_id      string
	Client_secret  string
	Get_secure_url int
	Grant_type     string
	Refresh_token  string
}

func NewClient(id, secret string) *Client {
	c := new(Client)
	c.clientId = id
	c.clientSecret = secret
	c.client = new(http.Client)
	c.client.Transport = cfbp.AddCloudFlareByPass(nil, cfbp.Options{
		Headers: map[string]string{
			"app-os":          "ios",
			"app-os-version":  ios_version,
			"User-Agent":      user_agent,
			"Accept-Language": "en-US,en;q=0.5",
		},
	})
	return c
}

func (p *Client) SetToken(access, refresh string) {
	p.accessToken = access
	p.refreshToken = refresh
}

func (p *Client) Login(token string) error {
	ts := time.Now().UTC().Format("2006-01-02T15:04:05") + "+00:00"
	data := authRequest{Client_id: p.clientId, Client_secret: p.clientSecret, Get_secure_url: 1, Refresh_token: token, Grant_type: "refresh_token"}
	mData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(mData)
	req, _ := http.NewRequest("POST", auth_url, buf)
	req.Header.Add("X-Client-Time", ts)
	sum := md5.Sum([]byte(ts + login_secret))
	req.Header.Add("X-Client-Hash", hex.EncodeToString(sum[:]))
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	var output struct {
		Response struct {
			User          User
			Access_token  string
			Refresh_token string
		}
		Error_description string
	}
	mData, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(mData))
	err = json.Unmarshal(mData, &output)
	if err != nil {
		return err
	}
	if output.Error_description != "" {
		return errors.New(output.Error_description)
	}
	p.accessToken = output.Response.Access_token
	p.refreshToken = output.Response.Refresh_token
	// TODO: Do something with user or comment it out
	return nil
}

func (p *Client) RefreshAuth() error {
	return p.Login(p.refreshToken)
}

func (p *Client) RefreshToken() string {
	return p.refreshToken
}

func (p *Client) buildGetRequest(url string) *http.Request {
	if p.accessToken == "" {
		return nil
	}
	req, _ := http.NewRequest("GET", base_url+url, http.NoBody)
	req.Header.Add("Authorization", "Bearer "+p.accessToken)
	return req
}

func (p *Client) buildPostRequest(url string, body io.Reader) *http.Request {
	if p.accessToken == "" {
		return nil
	}
	req, _ := http.NewRequest("POST", base_url+url, body)
	req.Header.Add("Authorization", "Bearer "+p.accessToken)
	return req
}

func (p *Client) GetIllust(id int) (*Illustration, error) {
	req := p.buildGetRequest("illust/detail?illust_id=" + strconv.Itoa(id))
	resp, err := p.client.Do(req)
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
		Illust Illustration
		Error  struct {
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
	output.Illust.client = p
	output.Illust.User.client = p
	return &output.Illust, nil
}

func (p *Client) GetUser(id int) (*User, error) {
	req := p.buildGetRequest("user/detail?filter=for_ios&user_id=" + strconv.Itoa(id))
	resp, err := p.client.Do(req)
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
		User User
		// Profile, Profile_publicity, Workspace
		Error struct {
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
	output.User.client = p
	return &output.User, nil
}

func (p *Client) RecommendedIllust(kind IllustrationType) (*IllustrationListing, error) {
	// TODO: Figure out how to make this call with NONE
	return p.newIllustrationListing("illust/recommended?content_type=" + string(kind))
}

func (p *Client) RankedIllust(mode RankingMode, day time.Time) (*IllustrationListing, error) {
	b := new(strings.Builder)
	b.WriteString("illust/ranking?mode=")
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
	b.WriteString("search/illust?word=")
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
	b.WriteString("search/user?word=")
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

// TODO: func (p *Client) GetUserMe() (*User, error) {}
