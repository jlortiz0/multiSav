package pixivapi_test

import (
	"testing"

	"github.com/jlortiz0/multisav/pixivapi"
)

func TestUserFetch(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	err := ret.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

func TestUserBookmarkTags(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	tags, err := ret.BookmarkTags(pixivapi.VISI_PUBLIC, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(tags.Bookmark_tags), tags.Offset)
	for _, x := range tags.Bookmark_tags {
		t.Log(x.Name, x.Count)
	}
}

func TestUserFollow(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	err := ret.Follow(pixivapi.VISI_PUBLIC)
	if err != nil {
		t.Fatal(err)
	}
	err = ret.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if !ret.Is_followed {
		t.Fatal("user was not followed")
	}
	err = ret.Unfollow()
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserFollowing(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	ls, err := ret.Following(pixivapi.VISI_PUBLIC)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Name)
	}
}

func TestUserFollowers(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	ls, err := ret.Followers()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Name)
	}
}

func TestUserRelated(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	ls, err := ret.Related()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Name)
	}
}

func TestUserIllustrations(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	ls, err := ret.Illustrations(pixivapi.ILTYPE_NONE)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Title)
	}
}

func TestUserBookmarks(t *testing.T) {
	p := loginHelper(t)
	ret := p.UserFromID(16944635)
	ls, err := ret.Bookmarks("", pixivapi.VISI_PUBLIC)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Log(err)
		}
		t.Log(n.ID, n.Title)
	}
}
