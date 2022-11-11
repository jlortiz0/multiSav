package redditapi_test

import "testing"

func TestRedditor(t *testing.T) {
	red := loginHelper(t)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		t.Fatalf("failed to load profile: %s", err.Error())
	}
	t.Log(user)
}

func TestSubmissions(t *testing.T) {
	red := loginHelper(t)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		t.Fatalf("failed to load profile: %s", err.Error())
	}
	iter, err := user.ListSubmissions(5)
	if err != nil {
		t.Fatal(err)
	}
	for iter.HasNext() {
		x, err := iter.Next()
		if err != nil {
			t.Error(err)
			break
		}
		t.Log(x.Title)
	}
}

func TestComments(t *testing.T) {
	red := loginHelper(t)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		t.Fatalf("failed to load profile: %s", err.Error())
	}
	iter, err := user.ListComments(5)
	if err != nil {
		t.Fatal(err)
	}
	for iter.HasNext() {
		x, err := iter.Next()
		if err != nil {
			t.Error(err)
			break
		}
		if len(x.Body) > 60 {
			x.Body = x.Body[:60]
		}
		t.Log(x.Body)
		if !x.Edited.IsZero() {
			t.Log(x.Edited)
		}
	}
}
