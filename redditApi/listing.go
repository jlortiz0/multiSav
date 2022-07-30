package redditapi

import (
	"encoding/json"
	"fmt"
	"io"
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

func newSubmissionIterator(URL string, red *Reddit, data []byte, limit uint32) (*SubmissionIterator, error) {
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
		url := fmt.Sprintf("%s?after=%s&count=%d&limit=%d", iter.URL, iter.lastId, iter.count, LISTING_PAGE_LIMIT)
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
	return (iter.limit != 0 && iter.count < iter.limit) || iter.lastId != ""
}

func (iter *SubmissionIterator) Count() uint32 {
	return iter.count
}
