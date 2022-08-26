package redditapi_test

import "testing"

func TestMultiredditSlice(T *testing.T) {
	red := loginHelper(T)
	multis, err := red.Self().Multireddits()
	if err != nil {
		T.Fatal(err)
	}
	for _, x := range multis {
		T.Log(x.Name)
		if x.Name == "" {
			T.Fail()
		}
	}
	user, err := red.Redditor("midnightrazorheart")
	if err != nil {
		T.Fatal(err)
	}
	multis, err = user.Multireddits()
	if err != nil {
		T.Fatal(err)
	}
	for _, x := range multis {
		T.Log(x.Name)
		if x.Name == "" {
			T.Fail()
		}
	}
	red.Logout()
}
