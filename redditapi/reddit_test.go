package redditapi_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/jlortiz0/multisav/redditapi"
)

func TestLogin(t *testing.T) {
	loginHelper(t)
}

func loginHelper(t *testing.T) *redditapi.Reddit {
	t.Helper()
	data := make([]byte, 1024)
	f, err := os.Open("login.json")
	if err != nil {
		t.Fatalf("Failed to open login data file: %s", err.Error())
	}
	n, err := f.Read(data)
	if err != nil {
		t.Fatalf("Failed to read login data: %s", err.Error())
	}
	f.Close()
	var fields struct {
		Id      string
		Secret  string
		Refresh string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		t.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := redditapi.NewReddit("linux:org.jlortiz.test.GolangRedditAPI:v0.0.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Refresh)
	if err != nil {
		t.Fatalf("Failed to log in: %s", err.Error())
	}
	return red
}

func TestListingNew(t *testing.T) {
	red := loginHelper(t)
	ls, err := red.ListNew(0)
	if err != nil {
		t.Fatalf("Failed to get /new: %s", err.Error())
	}
	if !ls.HasNext() {
		t.Fatal("Listing should not be empty")
	}
	for i := 0; i < 5; i++ {
		x, err := ls.Next()
		if err != nil {
			t.Error(err.Error())
		} else if x == nil {
			if ls.Count() != 4 {
				t.Error("Ended before the end?")
			}
		} else {
			t.Log(x.ID)
		}
	}
	if ls.Count() != 5 {
		t.Error("Listing count should be the number of things processed")
	}
}
