package redditapi_test

import (
	"testing"

	redditapi "jlortiz.org/rediSav/redditApi"
)

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
