package pixivapi_test

import (
	"testing"

	"jlortiz.org/redisav/pixivapi"
)

func TestIllustFetch(T *testing.T) {
	p := loginHelper(T)
	ret := p.IllustFromID(0)
	err := ret.Fetch()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ret)
	ret2, err := p.GetIllust(0)
	if err != nil {
		T.Fatal(err)
	}
	if ret.Title != ret2.Title {
		T.Fail()
	}
	T.Log(ret2)
}

func TestIllustBookmark(T *testing.T) {
	p := loginHelper(T)
	ret := p.IllustFromID(0)
	err := ret.Bookmark(pixivapi.VISI_PRIVATE)
	if err != nil {
		T.Fatal(err)
	}
	// How do we get ourselves again?
	// b, err := p.UserFromID(-1).Bookmarks("", pixivapi.VISI_PRIVATE)
	err = ret.Fetch()
	if err != nil {
		T.Fatal(err)
	}
	if !ret.Is_bookmarked {
		T.Fatal("image does not seem to be bookmarked")
	}
	err = ret.Unbookmark()
	if err != nil {
		T.Fatal(err)
	}
}

func TestUgoiraMeta(T *testing.T) {
	p := loginHelper(T)
	ret := p.IllustFromID(0)
	meta, err := ret.GetUgoiraMetadata()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(meta.Zip_urls)
	T.Log(len(meta.Frames))
}

func TestIllustComments(T *testing.T) {
	p := loginHelper(T)
	ret := p.IllustFromID(0)
	com, err := ret.GetComments(0)
	if err != nil {
		T.Fatal(err)
	}
	if len(com.Comments) == 0 {
		T.Error("No comments?")
	}
	T.Log(com.Offset, com.Total_comments, len(com.Comments))
	for _, x := range com.Comments {
		T.Log(x.Date, x.Comment, x.User.Name)
	}
}

func TestIllustRelated(T *testing.T) {
	p := loginHelper(T)
	ret := p.IllustFromID(0)
	ls, err := ret.GetRelated()
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			T.Error(err)
		}
		T.Log(n.ID, n.Title)
	}
}