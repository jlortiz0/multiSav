package redditapi_test

import "testing"

func TestRedditor(T *testing.T) {
	red := loginHelper(T)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		T.Fatalf("failed to load profile: %s", err.Error())
	}
	T.Log(user)
}

func TestSubmissions(T *testing.T) {
	red := loginHelper(T)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		T.Fatalf("failed to load profile: %s", err.Error())
	}
	iter, err := user.ListSubmissions(5)
	if err != nil {
		T.Fatal(err)
	}
	for iter.HasNext() {
		x, err := iter.Next()
		if err != nil {
			T.Error(err)
			break
		}
		T.Log(x.Title)
	}
}

func TestComments(T *testing.T) {
	red := loginHelper(T)
	user, err := red.Redditor("ketralnis")
	if err != nil {
		T.Fatalf("failed to load profile: %s", err.Error())
	}
	iter, err := user.ListComments(5)
	if err != nil {
		T.Fatal(err)
	}
	for iter.HasNext() {
		x, err := iter.Next()
		if err != nil {
			T.Error(err)
			break
		}
		if len(x.Body) > 60 {
			x.Body = x.Body[:60]
		}
		T.Log(x.Body)
		if !x.Edited.IsZero() {
			T.Log(x.Edited)
		}
	}
}
