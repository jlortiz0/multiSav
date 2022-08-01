package redditapi_test

import (
	"testing"

	redditapi "jlortiz.org/rediSav/redditApi"
)

func TestListNew(T *testing.T) {
	red := loginHelper(T)
	s, err := redditapi.NewSubreddit(red, "MisreadSprites")
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
