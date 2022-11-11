package redditapi_test

import (
	"testing"
)

func TestSubmissionSave(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Submission("b8yd3r")
	if err != nil {
		t.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Saved == true {
		t.Log("submission is already saved")
		t.SkipNow()
	}
	err = s.Save()
	if err != nil {
		t.Fatalf("failed to save: %s", err.Error())
	}
	if s.Saved == false {
		t.Error("submission object was not marked as saved")
	}
	s, err = red.Submission("b8yd3r")
	if err != nil {
		t.Fatalf("failed to load submission 2nd time: %s", err.Error())
	}
	if s.Saved == false {
		t.Error("submission not saved serverside")
	}
}

func TestSubmissionUnsave(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Submission("b8yd3r")
	if err != nil {
		t.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Saved == false {
		t.Log("submission is not saved")
		t.SkipNow()
	}
	err = s.Unsave()
	if err != nil {
		t.Fatalf("failed to unsave: %s", err.Error())
	}
	if s.Saved == true {
		t.Error("submission object was not unmarked as saved")
	}
	s, err = red.Submission("b8yd3r")
	if err != nil {
		t.Fatalf("failed to load submission 2nd time: %s", err.Error())
	}
	if s.Saved == true {
		t.Error("submission not unsaved serverside")
	}
}

func TestSubmissionVarious(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Submission("")
	if err != nil {
		t.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Author != red.Self().Name {
		t.Fatal("own post required")
	}
	err = s.Edit("Did you know that I am secretly a rumduck?")
	if err != nil {
		t.Errorf("failed to edit: %s", err.Error())
	}
	err = s.Reply("It's true.")
	if err != nil {
		t.Errorf("failed to reply: %s", err.Error())
	}
	err = s.Downvote()
	if err != nil {
		t.Errorf("failed to downvote: %s", err.Error())
	}
}

func TestSubmissionDelete(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Submission("")
	if err != nil {
		t.Fatalf("failed to load submission: %s", err.Error())
	}
	if s.Author != red.Self().Name {
		t.Fatal("own post required")
	}
	err = s.Delete()
	if err != nil {
		t.Errorf("failed to delete: %s", err.Error())
	}
}
