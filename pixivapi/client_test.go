package pixivapi_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"jlortiz.org/redisav/pixivapi"
)

func TestLogin(T *testing.T) {
	loginHelper(T)
}

func loginHelper(T *testing.T) *pixivapi.Client {
	T.Helper()
	data := make([]byte, 1024)
	f, err := os.Open("../redditapi/login.json")
	if err != nil {
		T.Fatalf("Failed to open login data file: %s", err.Error())
	}
	n, err := f.Read(data)
	if err != nil {
		T.Fatalf("Failed to read login data: %s", err.Error())
	}
	f.Close()
	var fields struct {
		PixivId     string
		PixivSecret string
		PixivToken  string
		PixivAccess string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		T.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := pixivapi.NewClient(fields.PixivId, fields.PixivSecret)
	if fields.PixivAccess != "" {
		red.SetToken(fields.PixivAccess, fields.PixivToken)
	} else {
		err = red.Login(fields.PixivToken)
		if err != nil {
			T.Fatalf("Failed to log in: %s", err.Error())
		}
		T.Log("New refresh token: " + red.RefreshToken())
	}
	return red
}

func TestIllust(T *testing.T) {
	p := loginHelper(T)
	ret, err := p.GetIllust(101469224)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ret)
}

func TestUser(T *testing.T) {
	p := loginHelper(T)
	ret, err := p.GetUser(0)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ret)
}

func TestRecommended(T *testing.T) {
	p := loginHelper(T)
	ls, err := p.RecommendedIllust(pixivapi.ILTYPE_ILUST)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Title)
	}
}

func TestRanked(T *testing.T) {
	p := loginHelper(T)
	ls, err := p.RankedIllust(pixivapi.DAY, time.Time{})
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Title)
	}
}

func TestSearchIllustrations(T *testing.T) {
	p := loginHelper(T)
	ls, err := p.SearchIllust("TERMS", pixivapi.TAGS_EXACT, pixivapi.DATE_DESC, pixivapi.WITHIN_NONE)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Title)
	}
}

func TestSearchUser(T *testing.T) {
	p := loginHelper(T)
	ls, err := p.SearchUser("TERMS", pixivapi.DATE_DESC, pixivapi.WITHIN_NONE)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Name)
	}
}
