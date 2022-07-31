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
	if s == nil || s.Name == "" {
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

func TestSubmissionVarious(T *testing.T) {
	red := loginHelper(T)
	s := redditapi.NewSubmission(red, "")
	if s == nil || s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission")
	}
	if s.Author != red.Self().Name {
		T.Fatal("Own post required")
	}
	err := s.Edit("Did you know that I am secretly a rumduck?")
	if err != nil {
		T.Error(err.Error())
	}
	err = s.Reply("It's true.")
	if err != nil {
		T.Error(err.Error())
	}
	err = s.Downvote()
	if err != nil {
		T.Error(err.Error())
	}
	red.Logout()
}

func TestSubmissionDelete(T *testing.T) {
	red := loginHelper(T)
	s := redditapi.NewSubmission(red, "")
	if s == nil || s.Name == "" {
		red.Logout()
		T.Fatal("Failed to load submission")
	}
	if s.Author != red.Self().Name {
		T.Fatal("Own post required")
	}
	err := s.Delete()
	if err != nil {
		T.Error(err.Error())
	}
	red.Logout()
}
