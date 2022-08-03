package main

import "jlortiz.org/redisav/redditapi"

type LazySubmissionList struct {
	iter *redditapi.SubmissionIterator
	data []*redditapi.Submission
}

func NewLazySubmissionList(iter *redditapi.SubmissionIterator) *LazySubmissionList {
	ls := new(LazySubmissionList)
	ls.iter = iter
	ls.data = make([]*redditapi.Submission, 0, 100)
	return ls
}

func (ls *LazySubmissionList) IsLazy() bool {
	if ls.iter == nil {
		return false
	}
	if !ls.iter.HasNext() {
		ls.iter = nil
		return false
	}
	return true
}

func (ls *LazySubmissionList) fetchNext() error {
	if !ls.IsLazy() {
		return nil
	}
	for ls.iter.HasNext() {
		x, err := ls.iter.Next()
		if err != nil {
			return err
		}
		if x == nil {
			break
		}
		ls.data = append(ls.data, x)
		if ls.iter.NextRequiresFetch() {
			break
		}
	}
	if !ls.iter.HasNext() {
		ls.iter = nil
	}
	return nil
}

func (ls *LazySubmissionList) Get(index int) *redditapi.Submission {
	for index >= len(ls.data) {
		if !ls.IsLazy() {
			return nil
		}
		ls.fetchNext()
	}
	return ls.data[index]
}

func (ls *LazySubmissionList) Len() int {
	return len(ls.data)
}
