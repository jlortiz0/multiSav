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
	"time"

	cfbp "github.com/DaRealFreak/cloudflare-bp-go"
)

const login_secret = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"
const auth_url = "https://oauth.secure.pixiv.net/auth/token"
const base_url = "https://app-api.pixiv.net/v1/"
const ios_version = "15.7"
const user_agent = "PixivIOSApp/7.15.11 (iOS " + ios_version + "; iPhone13,2)"

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

func (p *Client) GetIllust(id string) (*Illustration, error) {
	req := p.buildGetRequest("illust/detail?illust_id=" + id)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	f, _ := os.OpenFile("out.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	f.Write(data)
	f.Close()
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
