package redditapi_test

import (
	"encoding/json"
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
		T.Fatal("Listing should not be empty")
	}
	for i := 0; i < 5; i++ {
		x, err := ls.Next()
		if err != nil {
			T.Error(err.Error())
		} else {
			T.Log(x.ID)
		}
	}
	if ls.Count() != 5 {
		T.Error("Listing count should be the number of things processed")
	}
	red.Logout()
}

func TestSubmissionSave(T *testing.T) {
	red := loginHelper(T)
	s := redditapi.NewSubmission(red, "b8yd3r")
	if s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission")
	}
	if s.Saved == true {
		T.Log("Submission is already saved")
		T.SkipNow()
	}
	err := s.Save()
	if err != nil {
		T.Fatal(err.Error())
	}
	if s.Saved == false {
		T.Error("Submission object was not marked as saved")
	}
	s = redditapi.NewSubmission(red, "b8yd3r")
	if s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission 2nd try")
	}
	if s.Saved == false {
		T.Error("Submission not saved serverside")
	}
	red.Logout()
}

func TestSubmissionUnsave(T *testing.T) {
	red := loginHelper(T)
	s := redditapi.NewSubmission(red, "b8yd3r")
	if s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission")
	}
	if s.Saved == false {
		T.Log("Submission is not saved")
		T.SkipNow()
	}
	err := s.Unsave()
	if err != nil {
		T.Fatal(err.Error())
	}
	if s.Saved == true {
		T.Error("Submission object was not unmarked as saved")
	}
	s = redditapi.NewSubmission(red, "b8yd3r")
	if s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission 2nd try")
	}
	if s.Saved == true {
		T.Error("Submission not unsaved serverside")
	}
	red.Logout()
}
