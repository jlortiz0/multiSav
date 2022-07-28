package redditapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

var nilReader *strings.Reader = strings.NewReader("")

type Reddit struct {
	Client *http.Client
	// Will not be zero when logged in
	token        string
	tokenExpiry  time.Time
	refreshToken string
	clientId     string
	clientSecret string
	userAgent    string
}

func (r *Reddit) checkToken() error {
	if !r.tokenExpiry.IsZero() && r.tokenExpiry.Before(time.Now()) {
		buf := new(bytes.Buffer)
		buf.WriteString("grant_type=refresh_token&refresh_token=")
		buf.WriteString(r.refreshToken)
		req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", buf)
		req.SetBasicAuth(r.clientId, r.clientSecret)
		req.Header.Add("User-Agent", r.userAgent)
		resp, err := r.Client.Do(req)
		if err != nil {
			return err
		}
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(string(data))
		}
		var tokenData struct {
			Access_token string
			Expires_in   int
			Token_type   string
		}
		err = json.Unmarshal(data, &tokenData)
		if err != nil {
			return err
		}
		r.tokenExpiry = time.Now().Add(time.Duration(tokenData.Expires_in) * time.Second)
		r.token = tokenData.Token_type + " " + tokenData.Access_token
	}
	return nil
}

func (r *Reddit) Login(username, password string) error {
	buf := new(bytes.Buffer)
	buf.WriteString("grant_type=password&username")
	buf.WriteString(username)
	buf.WriteString("&password=")
	buf.WriteString(password)
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", buf)
	req.SetBasicAuth(r.clientId, r.clientSecret)
	req.Header.Add("User-Agent", r.userAgent)
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(string(data))
	}
	var tokenData struct {
		Access_token  string
		Expires_in    int
		Token_type    string
		Refresh_token string
	}
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return err
	}
	r.tokenExpiry = time.Now().Add(time.Duration(tokenData.Expires_in) * time.Second)
	r.token = tokenData.Token_type + " " + tokenData.Access_token
	r.refreshToken = tokenData.Refresh_token
	return nil
}

func (r *Reddit) Logout() {
	if r.token != "" {
		r.tokenExpiry = time.Time{}
		r.refreshToken = ""
		r.token = ""
	}
}

func (r *Reddit) buildRequest(method, url string, body io.Reader) *http.Request {
	if r.token != "" {
		url = "https://oauth.reddit.com/" + url
	} else {
		url = "https://reddit.com/" + url
	}
	rq, _ := http.NewRequest(method, url, body)
	// rq.SetBasicAuth(r.clientId, r.clientSecret)
	rq.Header.Add("User-Agent", r.userAgent)
	if r.token != "" {
		rq.Header.Add("Authorization", r.token)
	}
	return rq
}
