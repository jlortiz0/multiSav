package redditapi_test

import (
	"testing"

	redditapi "jlortiz.org/rediSav/redditApi"
)

func TestListNew(T *testing.T) {
	red := loginHelper(T)
	s, err := redditapi.NewSubreddit(red, "CountOnceADay")
	if err != nil {
		T.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.ListNew(5)
	if err != nil {
		T.Fatalf("failed to get sr/new: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		T.Log(post.Title)
		if iter.Count() > 5 {
			T.Error("iterator did not stop after 5")
			break
		}
	}
	red.Logout()
}

func TestListTop(T *testing.T) {
	red := loginHelper(T)
	s, err := redditapi.NewSubreddit(red, "CountOnceADay")
	if err != nil {
		T.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.ListTop(5, "")
	if err != nil {
		T.Fatalf("failed to get sr/top: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		T.Log(post.Title)
	}
	iter, err = s.ListTop(5, "week")
	if err != nil {
		T.Fatalf("failed to get sr/top?week: %s", err.Error())
	}
	T.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		T.Log(post.Title)
	}
	iter, err = s.ListTop(5, "all")
	if err != nil {
		T.Fatalf("failed to get sr/top: %s", err.Error())
	}
	T.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing?all: %s", err)
		}
		T.Log(post.Title)
	}
	red.Logout()
}

func TestSearch(T *testing.T) {
	red := loginHelper(T)
	s, err := redditapi.NewSubreddit(red, "MisreadSprites")
	if err != nil {
		T.Fatalf("failed to load subreddit: %s", err.Error())
	}
	iter, err := s.Search(5, "link", "", "")
	if err != nil {
		T.Fatalf("failed to get sr/search?link: %s", err.Error())
	}
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		T.Log(post.Title)
		if iter.Count() > 5 {
			T.Error("iterator did not stop after 5")
			break
		}
	}
	iter, err = s.Search(5, "link", "top", "")
	if err != nil {
		T.Fatalf("failed to get sr/search?link&top: %s", err.Error())
	}
	T.Log("---")
	for iter.HasNext() {
		post, err := iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		T.Log(post.Title)
		post, err = iter.Next()
		if err != nil {
			T.Errorf("failed to load submission from listing: %s", err)
		}
		if post != nil {
			T.Log(post.Title)
		}
		if iter.Count() > 5 {
			T.Error("iterator did not stop after 5")
			break
		}
	}
	red.Logout()
}

func TestSubscribe(T *testing.T) {
	red := loginHelper(T)
	s, err := redditapi.NewSubreddit(red, "CountOnceADay")
	if err != nil {
		T.Fatalf("failed to load subreddit: %s", err.Error())
	}
	if s.User_is_subscriber == true {
		T.Fatal("already subscribed to this subreddit")
	}
	err = s.Subscribe()
	if err != nil {
		T.Fatalf("failed to subscribe: %s", err.Error())
	}
	s, err = redditapi.NewSubreddit(red, "CountOnceADay")
	if err != nil {
		T.Fatalf("failed to load subreddit 2nd time: %s", err.Error())
	}
	if s.User_is_subscriber == false {
		T.Fatal("failed to subscribe serverside")
	}
	err = s.Unsubscribe()
	if err != nil {
		T.Fatalf("failed to unsubscribe: %s", err.Error())
	}
	red.Logout()
}
