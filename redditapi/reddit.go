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

func NewReddit(userAgent string, clientId, clientSecret string) *Reddit {
	red := new(Reddit)
	red.Client = new(http.Client)
	red.userAgent = userAgent
	red.clientId = clientId
	red.clientSecret = clientSecret
	return red
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
	buf.WriteString("grant_type=password&username=")
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
		Error         string
	}
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return err
	}
	if tokenData.Error != "" {
		return errors.New(tokenData.Error)
	}
	r.tokenExpiry = time.Now().Add(time.Duration(tokenData.Expires_in) * time.Second)
	r.token = tokenData.Token_type + " " + tokenData.Access_token
	r.refreshToken = tokenData.Refresh_token
	return nil
}

func (r *Reddit) Logout() error {
	if r.token != "" {
		buf := new(bytes.Buffer)
		buf.WriteString("token=")
		buf.WriteString(r.refreshToken)
		buf.WriteString("&token_type_hint=refresh_token")
		req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/revoke_token", buf)
		req.SetBasicAuth(r.clientId, r.clientSecret)
		req.Header.Add("User-Agent", r.userAgent)
		resp, err := r.Client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 204 {
			data, _ := io.ReadAll(resp.Body)
			return errors.New(string(data))
		}
		r.tokenExpiry = time.Time{}
		r.refreshToken = ""
		r.token = ""
	}
	return nil
}

func (r *Reddit) buildRequest(method, url string, body io.Reader) *http.Request {
	if r.token != "" {
		r.checkToken()
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

func (r *Reddit) ListNew(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("new", r, limit)
}

func (r *Reddit) Self() *Redditor {
	rq := r.buildRequest("GET", "api/v1/me", nilReader)
	resp, err := r.Client.Do(rq)
	if err != nil {
		return nil
	}
	data, _ := io.ReadAll(resp.Body)
	var user Redditor
	json.Unmarshal(data, &user)
	return &user
}

type Timestamp struct{ time.Time }

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var ts float64
	err := json.Unmarshal(b, &ts)
	if err != nil {
		return err
	}
	*t = Timestamp{time.Unix(int64(ts), 0)}
	return nil
}

type TSBool struct{ time.Time }

func (t *TSBool) UnmarshalJSON(b []byte) error {
	var ts float64
	if b[0] != 'f' {
		err := json.Unmarshal(b, &ts)
		if err != nil {
			return err
		}
		*t = TSBool{time.Unix(int64(ts), 0)}
	} else {
		*t = TSBool{}
	}
	return nil
}
