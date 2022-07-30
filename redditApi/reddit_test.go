package redditapi_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	redditapi "jlortiz.org/rediSav/redditApi"
)

func TestLogin(T *testing.T) {
	red := loginHelper(T)
	red.Logout()
}

func loginHelper(T *testing.T) *redditapi.Reddit {
	T.Helper()
	data := make([]byte, 256)
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
		Id       string
		Secret   string
		Login    string
		Password string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		T.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := redditapi.NewReddit("linux:org.jlortiz.test.GolangRedditAPI:v0.0.1 (by /u/jlortiz)", fields.Id, fields.Secret)
	if red == nil {
		T.FailNow()
	}
	err = red.Login(fields.Login, fields.Password)
	if err != nil {
		T.Fatalf("Failed to log in: %s", err.Error())
	}
	return red
}

func TestListingNew(T *testing.T) {
	red := loginHelper(T)
	ls, err := red.GetNew()
	if err != nil {
		T.Fatalf("Failed to get /new: %s", err.Error())
	}
	if !ls.HasNext() {
		T.Log("Listing should not be empty")
		T.Fail()
	}
	x, err := ls.NextMap()
	fmt.Println(err)
	for k, v := range x {
		fmt.Println(k, v)
	}
	if ls.Len() != 1 {
		T.Log("Listing length should be 1")
		T.Fail()
	}
	ls.Close()
	red.Logout()
}
