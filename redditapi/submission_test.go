package redditapi_test

import (
	"testing"
)

func TestSubmissionSave(T *testing.T) {
	red := loginHelper(T)
	s, err := red.Submission("b8yd3r")
	if err != nil {
		T.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Saved == true {
		T.Log("submission is already saved")
		T.SkipNow()
	}
	err = s.Save()
	if err != nil {
		T.Fatalf("failed to save: %s", err.Error())
	}
	if s.Saved == false {
		T.Error("submission object was not marked as saved")
	}
	s, err = red.Submission("b8yd3r")
	if err != nil {
		T.Fatalf("failed to load submission 2nd time: %s", err.Error())
	}
	if s.Saved == false {
		T.Error("submission not saved serverside")
	}
}

func TestSubmissionUnsave(T *testing.T) {
	red := loginHelper(T)
	s, err := red.Submission("b8yd3r")
	if err != nil {
		T.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Saved == false {
		T.Log("submission is not saved")
		T.SkipNow()
	}
	err = s.Unsave()
	if err != nil {
		T.Fatalf("failed to unsave: %s", err.Error())
	}
	if s.Saved == true {
		T.Error("submission object was not unmarked as saved")
	}
	s, err = red.Submission("b8yd3r")
	if err != nil {
		T.Fatalf("failed to load submission 2nd time: %s", err.Error())
	}
	if s.Saved == true {
		T.Error("submission not unsaved serverside")
	}
}

func TestSubmissionVarious(T *testing.T) {
	red := loginHelper(T)
	s, err := red.Submission("")
	if err != nil {
		T.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Author != red.Self().Name {
		T.Fatal("own post required")
	}
	err = s.Edit("Did you know that I am secretly a rumduck?")
	if err != nil {
		T.Errorf("failed to edit: %s", err.Error())
	}
	err = s.Reply("It's true.")
	if err != nil {
		T.Errorf("failed to reply: %s", err.Error())
	}
	err = s.Downvote()
	if err != nil {
		T.Errorf("failed to downvote: %s", err.Error())
	}
}

func TestSubmissionDelete(T *testing.T) {
	red := loginHelper(T)
	s, err := red.Submission("")
	if err != nil {
		T.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Author != red.Self().Name {
		T.Fatal("own post required")
	}
	err = s.Delete()
	if err != nil {
		T.Errorf("failed to delete: %s", err.Error())
	}
}
