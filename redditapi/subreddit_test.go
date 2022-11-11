package redditapi_test

import (
	"testing"
)

func TestListNew(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Subreddit("CountOnceADay")
	if err != nil {
		t.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.ListNew(5)
	if err != nil {
		t.Fatalf("failed to get sr/new: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		t.Log(post.Title)
		if iter.Count() > 5 {
			t.Error("iterator did not stop after 5")
			break
		}
	}
}

func TestListTop(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Subreddit("CountOnceADay")
	if err != nil {
		t.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.ListTop(5, "")
	if err != nil {
		t.Fatalf("failed to get sr/top: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		t.Log(post.Title)
	}
	iter, err = s.ListTop(5, "week")
	if err != nil {
		t.Fatalf("failed to get sr/top?week: %s", err.Error())
	}
	t.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		t.Log(post.Title)
	}
	iter, err = s.ListTop(5, "all")
	if err != nil {
		t.Fatalf("failed to get sr/top: %s", err.Error())
	}
	t.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing?all: %s", err)
		}
		t.Log(post.Title)
	}
}

func TestSearch(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Subreddit("MisreadSprites")
	if err != nil {
		t.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.Search(5, "link", "", "")
	if err != nil {
		t.Fatalf("failed to get sr/search?link: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		t.Log(post.Title)
		if iter.Count() > 5 {
			t.Error("iterator did not stop after 5")
			break
		}
	}
	iter, err = s.Search(5, "link", "top", "")
	if err != nil {
		t.Fatalf("failed to get sr/search?link&top: %s", err.Error())
	}
	t.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		t.Log(post.Title)
		post, err = iter.Next()
		if err != nil {
			t.Errorf("failed to load submission from listing: %s", err)
		}
		if post != nil {
			t.Log(post.Title)
		}
		if iter.Count() > 5 {
			t.Error("iterator did not stop after 5")
			break
		}
	}
}

func TestSubscribe(t *testing.T) {
	red := loginHelper(t)
	s, err := red.Subreddit("CountOnceADay")
	if err != nil {
		t.Fatalf("failed to load subreddit: %s", err.Error())
	}
	if s.User_is_subscriber == true {
		t.Fatal("already subscribed to this subreddit")
	}
	err = s.Subscribe()
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err.Error())
	}
	s, err = red.Subreddit("CountOnceADay")
	if err != nil {
		t.Fatalf("failed to load subreddit 2nd time: %s", err.Error())
	}
	if s.User_is_subscriber == false {
		t.Fatal("failed to subscribe serverside")
	}
	err = s.Unsubscribe()
	if err != nil {
		t.Fatalf("failed to unsubscribe: %s", err.Error())
	}
}
