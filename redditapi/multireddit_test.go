package redditapi_test

import "testing"

func TestMultiredditSlice(t *testing.T) {
	red := loginHelper(t)
	multis, err := red.Self().Multireddits()
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range multis {
		t.Log(x.Name)
		if x.Name == "" {
			t.Fail()
		}
	}
	user, err := red.Redditor("midnightrazorheart")
	if err != nil {
		t.Fatal(err)
	}
	multis, err = user.Multireddits()
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range multis {
		t.Log(x.Name)
		if x.Name == "" {
			t.Fail()
		}
	}
}
