package redditapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const LISTING_PAGE_LIMIT = 100

type SubmissionIterator struct {
	URL    string
	Reddit *Reddit
	count  int
	limit  int
	lastId string
	index  int
	data   []struct {
		Kind string
		Data *Submission
	}
}

type submissionListingPayload struct {
	Data struct {
		After    string
		Children []struct {
			Kind string
			Data *Submission
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func newSubmissionIterator(URL string, red *Reddit, limit int) (*SubmissionIterator, error) {
	req := red.buildRequest("GET", URL, nilReader)
	resp, err := red.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(string(data))
	}
	return newSubmissionIteratorPayload(URL, red, data, limit)
}

func newSubmissionIteratorPayload(URL string, red *Reddit, data []byte, limit int) (*SubmissionIterator, error) {
	i := new(SubmissionIterator)
	i.URL = URL
	i.limit = limit
	i.Reddit = red
	var payload submissionListingPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	i.lastId = payload.Data.After
	i.data = payload.Data.Children
	return i, nil
}

func (iter *SubmissionIterator) Next() (*Submission, error) {
	if !iter.HasNext() {
		return nil, nil
	}
	if iter.index == len(iter.data) {
		chr := '?'
		if strings.ContainsRune(iter.URL, '?') {
			chr = '&'
		}
		url := fmt.Sprintf("%s%cafter=%s&count=%d&limit=%d", iter.URL, chr, iter.lastId, iter.count, minInt(LISTING_PAGE_LIMIT, iter.limit-iter.count))
		resp, err := iter.Reddit.Client.Do(iter.Reddit.buildRequest("GET", url, nilReader))
		if err != nil {
			return nil, err
		}
		var payload submissionListingPayload
		data, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return nil, err
		}
		iter.lastId = payload.Data.After
		iter.data = payload.Data.Children
		iter.index = 0
		if len(iter.data) == 0 {
			return nil, nil
		}
	}
	iter.count++
	iter.index++
	if iter.data[iter.index-1].Kind != "t3" {
		return iter.Next()
	}
	return iter.data[iter.index-1].Data, nil
}

func (iter *SubmissionIterator) HasNext() bool {
	if iter.limit == 0 {
		return len(iter.data) != 0
	} else {
		return iter.count < iter.limit
	}
}

func (iter *SubmissionIterator) NextRequiresFetch() bool {
	return iter.index == len(iter.data)
}

func (iter *SubmissionIterator) Count() int {
	return iter.count
}

type CommentIterator struct {
	URL    string
	Reddit *Reddit
	count  int
	limit  int
	lastId string
	index  int
	data   []struct {
		Kind string
		Data *Comment
	}
}

type commentListingPayload struct {
	Data struct {
		After    string
		Children []struct {
			Kind string
			Data *Comment
		}
	}
}

func newCommentIterator(URL string, red *Reddit, limit int) (*CommentIterator, error) {
	req := red.buildRequest("GET", URL, nilReader)
	resp, err := red.Client.Do(req)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(string(data))
	}
	return newCommentIteratorPayload(URL, red, data, limit)
}

func newCommentIteratorPayload(URL string, red *Reddit, data []byte, limit int) (*CommentIterator, error) {
	i := new(CommentIterator)
	i.URL = URL
	i.limit = limit
	i.Reddit = red
	var payload commentListingPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return nil, err
	}
	i.lastId = payload.Data.After
	i.data = payload.Data.Children
	return i, nil
}

func (iter *CommentIterator) Next() (*Comment, error) {
	if !iter.HasNext() {
		return nil, nil
	}
	if iter.index == len(iter.data) {
		chr := '?'
		if strings.ContainsRune(iter.URL, '?') {
			chr = '&'
		}
		url := fmt.Sprintf("%s%cafter=%s&count=%d&limit=%d", iter.URL, chr, iter.lastId, iter.count, minInt(LISTING_PAGE_LIMIT, iter.limit-iter.count))
		resp, err := iter.Reddit.Client.Do(iter.Reddit.buildRequest("GET", url, nilReader))
		if err != nil {
			return nil, err
		}
		var payload commentListingPayload
		data, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(data, &payload)
		if err != nil {
			return nil, err
		}
		iter.lastId = payload.Data.After
		iter.data = payload.Data.Children
		iter.index = 0
		if len(iter.data) == 0 {
			return nil, nil
		}
	}
	iter.count++
	iter.index++
	if iter.data[iter.index-1].Kind != "t1" {
		return iter.Next()
	}
	return iter.data[iter.index-1].Data, nil
}

func (iter *CommentIterator) HasNext() bool {
	if iter.limit == 0 {
		return len(iter.data) != 0
	} else {
		return iter.count < iter.limit
	}
}

func (iter *CommentIterator) NextRequiresFetch() bool {
	return iter.index == len(iter.data)
}

func (iter *CommentIterator) Count() int {
	return iter.count
}
