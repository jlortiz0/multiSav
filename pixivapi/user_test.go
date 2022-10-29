package pixivapi_test

import (
	"testing"

	"github.com/jlortiz0/multisav/pixivapi"
)

func TestUserFetch(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	err := ret.Fetch()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ret)
}

func TestUserBookmarkTags(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	tags, err := ret.BookmarkTags(pixivapi.VISI_PUBLIC, 0)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(len(tags.Bookmark_tags), tags.Offset)
	for _, x := range tags.Bookmark_tags {
		T.Log(x.Name, x.Count)
	}
}

func TestUserFollow(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	err := ret.Follow(pixivapi.VISI_PUBLIC)
	if err != nil {
		T.Fatal(err)
	}
	err = ret.Fetch()
	if err != nil {
		T.Fatal(err)
	}
	if !ret.Is_followed {
		T.Fatal("user was not followed")
	}
	err = ret.Unfollow()
	if err != nil {
		T.Fatal(err)
	}
}

func TestUserFollowing(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	ls, err := ret.Following(pixivapi.VISI_PUBLIC)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Name)
	}
}

func TestUserFollowers(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	ls, err := ret.Followers()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Name)
	}
}

func TestUserRelated(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	ls, err := ret.Related()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Name)
	}
}

func TestUserIllustrations(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	ls, err := ret.Illustrations(pixivapi.ILTYPE_NONE)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Title)
	}
}

func TestUserBookmarks(T *testing.T) {
	p := loginHelper(T)
	ret := p.UserFromID(16944635)
	ls, err := ret.Bookmarks("", pixivapi.VISI_PUBLIC)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Log(err)
		}
		T.Log(n.ID, n.Title)
	}
}
