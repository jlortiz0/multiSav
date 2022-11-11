package pixivapi_test

import (
	"testing"

	"github.com/jlortiz0/multisav/pixivapi"
)

func TestIllustFetch(t *testing.T) {
	p := loginHelper(t)
	ret := p.IllustFromID(101471765)
	err := ret.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
	ret2, err := p.GetIllust(101471765)
	if err != nil {
		t.Fatal(err)
	}
	if ret.Title != ret2.Title {
		t.Fail()
	}
	t.Log(ret2)
}

func TestIllustBookmark(t *testing.T) {
	p := loginHelper(t)
	ret := p.IllustFromID(101471765)
	err := ret.Bookmark(pixivapi.VISI_PUBLIC)
	if err != nil {
		t.Fatal(err)
	}
	// How do we get ourselves again?
	// b, err := p.UserFromID(-1).Bookmarks("", pixivapi.VISI_PRIVATE)
	err = ret.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
	if !ret.Is_bookmarked {
		t.Fatal("image does not seem to be bookmarked")
	}
	err = ret.Unbookmark()
	if err != nil {
		t.Fatal(err)
	}
}

func TestUgoiraMeta(t *testing.T) {
	p := loginHelper(t)
	ret := p.IllustFromID(87063503)
	meta, err := ret.GetUgoiraMetadata()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(meta.Zip_urls)
	t.Log(len(meta.Frames))
}

func TestIllustComments(t *testing.T) {
	p := loginHelper(t)
	ret := p.IllustFromID(101490348)
	com, err := ret.GetComments(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(com.Comments) == 0 {
		t.Error("No comments?")
	}
	t.Log(com.Offset, com.Total_comments, len(com.Comments))
	for _, x := range com.Comments {
		t.Log(x.Date, x.Comment, x.User.Name)
	}
}

func TestIllustRelated(t *testing.T) {
	p := loginHelper(t)
	ret := p.IllustFromID(101471765)
	ls, err := ret.GetRelated()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ls.Buffered())
	for !ls.NextRequiresFetch() {
		n, err := ls.Next()
		if err != nil {
			t.Error(err)
		}
		t.Log(n.ID, n.Title)
	}
}
