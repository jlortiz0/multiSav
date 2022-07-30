package redditapi

import (
	"encoding/json"
	"fmt"
)

const LISTING_PAGE_LIMIT = "100"

type ListingIterator struct {
	URL     string
	Reddit  *Reddit
	count   uint32
	lastId  string
	decoder *json.Decoder
	Close   func() error
}

type Snowflake interface {
	GetId() string
}

type LimitedDecoder interface {
	Decode(interface{}) error
}

func (iter *ListingIterator) Next(parser func(LimitedDecoder) Snowflake) (Snowflake, error) {
	if !iter.decoder.More() {
		url := fmt.Sprintf("%s?after=%s&count=%d&limit="+LISTING_PAGE_LIMIT, iter.URL, iter.lastId, iter.count)
		resp, err := iter.Reddit.Client.Do(iter.Reddit.buildRequest("GET", url, nilReader))
		if err != nil {
			return nil, err
		}
		iter.decoder = json.NewDecoder(resp.Body)
		if !iter.decoder.More() {
			iter.decoder = nil
			return nil, nil
		}
	}
	data := parser(iter.decoder)
	iter.lastId = data.GetId()
	iter.count++
	return data, nil
}

type JsonObj map[string]interface{}

func (obj JsonObj) GetId() string {
	s := obj["id"]
	if s == nil {
		return ""
	}
	return s.(string)
}

func (iter *ListingIterator) NextMap() (JsonObj, error) {
	obj, err := iter.Next(func(ld LimitedDecoder) Snowflake {
		var data JsonObj
		ld.Decode(&data)
		return data
	})
	return obj.(JsonObj), err
}

func (iter *ListingIterator) HasNext() bool {
	return iter.decoder != nil
}

func (iter *ListingIterator) Len() uint32 {
	return iter.count
}
