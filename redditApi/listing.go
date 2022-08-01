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
	count  uint32
	limit  uint32
	lastId string
	index  int
	data   []*Submission
}

type submissionListingPayload struct {
	Data struct {
		After    string
		Children []struct {
			Data *Submission
		}
	}
}

func newSubmissionIterator(URL string, red *Reddit, limit uint32) (*SubmissionIterator, error) {
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

func newSubmissionIteratorPayload(URL string, red *Reddit, data []byte, limit uint32) (*SubmissionIterator, error) {
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
	i.data = make([]*Submission, len(payload.Data.Children))
	for k, v := range payload.Data.Children {
		i.data[k] = v.Data
	}
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
		url := fmt.Sprintf("%s%cafter=%s&count=%d&limit=%d", iter.URL, chr, iter.lastId, iter.count, LISTING_PAGE_LIMIT)
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
		iter.data = make([]*Submission, len(payload.Data.Children))
		for k, v := range payload.Data.Children {
			iter.data[k] = v.Data
		}
		iter.index = 0
	}
	iter.count++
	iter.index++
	return iter.data[iter.index-1], nil
}

func (iter *SubmissionIterator) HasNext() bool {
	if iter.limit == 0 {
		return iter.lastId != ""
	} else {
		return iter.count < iter.limit
	}
}

func (iter *SubmissionIterator) Count() uint32 {
	return iter.count
}
