package redditapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type Reddit struct {
	config    *oauth2.Config
	token     *oauth2.Token
	userAgent string
}

func NewReddit(userAgent string, clientId, clientSecret string) *Reddit {
	red := new(Reddit)
	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:5738/reddit",
		Scopes:       []string{"history", "identity", "read", "save", "subscribe"},
	}
	config.Endpoint.AuthURL = "https://www.reddit.com/api/v1/authorize"
	config.Endpoint.TokenURL = "https://www.reddit.com/api/v1/access_token"
	red.userAgent = userAgent
	red.config = config
	return red
}

func (r *Reddit) checkToken() {
	if r.token != nil && !r.token.Valid() {
		token, err := r.config.TokenSource(context.Background(), r.token).Token()
		if err != nil {
			return
		}
		r.token = token
	}
}

func (r *Reddit) Login(refresh string) error {
	token := &oauth2.Token{RefreshToken: refresh}
	token, err := r.config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return err
	}
	r.token = token
	return nil
}

func (r *Reddit) buildRequest(method, url string, body io.Reader) *http.Request {
	if r.token != nil {
		r.checkToken()
		url = "https://oauth.reddit.com/" + url
	} else {
		url = "https://reddit.com/" + url
	}
	rq, _ := http.NewRequest(method, url, body)
	rq.Header.Add("User-Agent", r.userAgent)
	r.token.SetAuthHeader(rq)
	return rq
}

func (r *Reddit) GetRequest(url string) (*http.Response, error) {
	rq, _ := http.NewRequest("GET", url, http.NoBody)
	rq.Header.Add("User-Agent", r.userAgent)
	// TODO: Is this needed?
	r.token.SetAuthHeader(rq)
	return http.DefaultClient.Do(rq)
}

func (r *Reddit) BySubmissionId(s []string, limit int) (*SubmissionIterator, error) {
	for i, x := range s {
		if x[2] != '_' {
			s[i] = "t3_" + x
		}
	}
	return newSubmissionIterator("by_id/"+strings.Join(s, ","), r, limit)
}

func (r *Reddit) ListNew(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("new", r, limit)
}

func (r *Reddit) ListHot(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("hot", r, limit)
}

func (r *Reddit) ListControversial(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("controversial", r, limit)
}

func (r *Reddit) ListRising(limit int) (*SubmissionIterator, error) {
	return newSubmissionIterator("rising", r, limit)
}

// If t not specified, seems to default to "day"
func (r *Reddit) ListTop(limit int, t string) (*SubmissionIterator, error) {
	s := "top"
	if t != "" {
		s += "?t=" + t
	}
	return newSubmissionIterator(s, r, limit)
}

func (r *Reddit) Search(limit int, q string, sort string, t string) (*SubmissionIterator, error) {
	s := "search?q=" + url.QueryEscape(q)
	if t != "" {
		s += "&t=" + t
	}
	if sort != "" {
		s += "&sort=" + sort
	}
	return newSubmissionIterator(s, r, limit)
}

func (r *Reddit) Self() *Redditor {
	rq := r.buildRequest("GET", "api/v1/me", http.NoBody)
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		return nil
	}
	data, _ := io.ReadAll(resp.Body)
	var user Redditor
	json.Unmarshal(data, &user)
	user.reddit = r
	user.self = true
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
