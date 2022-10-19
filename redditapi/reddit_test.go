package redditapi_test

import (
	"encoding/json"
	"os"
	"testing"

	"jlortiz.org/multisav/redditapi"
)

func TestLogin(T *testing.T) {
	loginHelper(T)
}

func loginHelper(T *testing.T) *redditapi.Reddit {
	T.Helper()
	data := make([]byte, 1024)
	f, err := os.Open("login.json")
	if err != nil {
		T.Fatalf("Failed to open login data file: %s", err.Error())
	}
	n, err := f.Read(data)
	if err != nil {
		T.Fatalf("Failed to read login data: %s", err.Error())
	}
	f.Close()
	var fields struct {
		Id      string
		Secret  string
		Refresh string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		T.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := redditapi.NewReddit("linux:org.jlortiz.test.GolangRedditAPI:v0.0.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	err = red.Login(fields.Refresh)
	if err != nil {
		T.Fatalf("Failed to log in: %s", err.Error())
	}
	return red
}

func TestListingNew(T *testing.T) {
	red := loginHelper(T)
	ls, err := red.ListNew(0)
	if err != nil {
		T.Fatalf("Failed to get /new: %s", err.Error())
	}
	if !ls.HasNext() {
		T.Fatal("Listing should not be empty")
	}
	for i := 0; i < 5; i++ {
		x, err := ls.Next()
		if err != nil {
			T.Error(err.Error())
		} else if x == nil {
			if ls.Count() != 4 {
				T.Error("Ended before the end?")
			}
		} else {
			T.Log(x.ID)
		}
	}
	if ls.Count() != 5 {
		T.Error("Listing count should be the number of things processed")
	}
}
