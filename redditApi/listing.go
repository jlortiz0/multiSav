package redditapi

import (
	"encoding/json"
	"fmt"
)

const LISTING_PAGE_LIMIT = "100"

type ListingIterator struct {
	URL     string
	Reddit  Reddit
	count   uint32
	lastId  string
	decoder *json.Decoder
}

func (iter *ListingIterator) Next() interface{} {
	if !iter.decoder.More() {
		url := fmt.Sprintf("%s?after=%s&count=%d&limit="+LISTING_PAGE_LIMIT, iter.URL, iter.lastId, iter.count)
		resp, err := iter.Reddit.Client.Do(iter.Reddit.buildRequest("GET", url, nilReader))
		if err != nil {
			return err
		}
		iter.decoder = json.NewDecoder(resp.Body)
		if !iter.decoder.More() {
			iter.decoder = nil
			return nil
		}
	}
	var data interface{}
	iter.decoder.Decode(data)
	//iter.lastId = data.Id
	iter.count++
	return data
}

func (iter *ListingIterator) HasNext() bool {
	return iter.decoder != nil
}

func (iter *ListingIterator) Len() uint32 {
	return iter.count
}
