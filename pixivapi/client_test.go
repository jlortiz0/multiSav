package pixivapi_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jlortiz0/multisav/pixivapi"
)

func TestLogin(t *testing.T) {
	loginHelper(t)
}

func loginHelper(t *testing.T) *pixivapi.Client {
	t.Helper()
	data := make([]byte, 1024)
	f, err := os.Open("../redditapi/login.json")
	if err != nil {
		t.Fatalf("Failed to open login data file: %s", err.Error())
	}
	n, err := f.Read(data)
	if err != nil {
		t.Fatalf("Failed to read login data: %s", err.Error())
	}
	f.Close()
	var fields struct {
		PixivToken string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		t.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := pixivapi.NewClient()
	err = red.Login(fields.PixivToken)
	if err != nil {
		t.Fatalf("Failed to log in: %s", err.Error())
	}
	return red
}

func TestIllust(t *testing.T) {
	p := loginHelper(t)
	ret, err := p.GetIllust(101469224)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

func TestUser(t *testing.T) {
	p := loginHelper(t)
	ret, err := p.GetUser(16944635)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

func TestRecommended(t *testing.T) {
	p := loginHelper(t)
	ls, err := p.RecommendedIllust(pixivapi.ILTYPE_ILUST)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Title)
	}
}

func TestRanked(t *testing.T) {
	p := loginHelper(t)
	ls, err := p.RankedIllust(pixivapi.DAY, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Title)
	}
}

func TestSearchIllustrations(t *testing.T) {
	p := loginHelper(t)
	ls, err := p.SearchIllust("ugoira", pixivapi.TAGS_EXACT, pixivapi.DATE_DESC, pixivapi.WITHIN_NONE)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Title)
	}
}

func TestSearchUser(t *testing.T) {
	p := loginHelper(t)
	ls, err := p.SearchUser("South_AC", pixivapi.DATE_DESC, pixivapi.WITHIN_NONE)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Name)
	}
}
