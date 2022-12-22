package main_test

import (
	"strings"
	"testing"

	"github.com/jlortiz0/multisav/redditapi"
)

func splitAny(s string, seps string) []string {
	out := make([]string, 0, len(s)*len(seps)/20+1)
	ind := strings.IndexAny(s, seps)
	for ind != -1 {
		if ind != 0 {
			out = append(out, s[:ind])
		}
		s = s[ind+1:]
		ind = strings.IndexAny(s, seps)
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

func splitAny2(s string, seps string) []string {
	fast := make(map[rune]struct{}, len(seps))
	for _, x := range seps {
		fast[x] = struct{}{}
	}
	out := make([]string, 0, len(s)*len(seps)/20+1)
	ind := strings.IndexFunc(s, func(r rune) bool { _, ok := fast[r]; return ok })
	for ind != -1 {
		if ind != 0 {
			out = append(out, s[:ind])
		}
		s = s[ind+1:]
		ind = strings.IndexFunc(s, func(r rune) bool { _, ok := fast[r]; return ok })
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

func splitAny3(s string, seps string) []string {
	fast := make([]bool, 256)
	for _, x := range seps {
		fast[x] = true
	}
	out := make([]string, 0, len(s)*len(seps)/20+1)
	ind := strings.IndexFunc(s, func(r rune) bool { return fast[r] })
	for ind != -1 {
		if ind != 0 {
			out = append(out, s[:ind])
		}
		s = s[ind+1:]
		ind = strings.IndexFunc(s, func(r rune) bool { return fast[r] })
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

func setupHelper(b *testing.B) string {
	b.Helper()
	red := redditapi.NewReddit("linux:org.jlortiz.multiSav:v0.7.0 (by /u/jlortiz)", RedditID, RedditSecret)
	sub, err := red.Submission("hposam")
	if err != nil {
		panic(err)
	}
	if sub.Selftext == "" {
		panic("no selftext")
	}
	b.ResetTimer()
	return sub.Selftext
}

func BenchmarkSplitAny(b *testing.B) {
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		splitAny(s, " \n\t()[]")
	}
}

func BenchmarkSplitFunc(b *testing.B) {
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		strings.FieldsFunc(s, func(c rune) bool { return strings.ContainsRune(" \n\t()[]", c) })
	}
}

func BenchmarkSplitFunc2(b *testing.B) {
	fast := make(map[rune]struct{}, 7)
	fast[' '] = struct{}{}
	fast['('] = struct{}{}
	fast[')'] = struct{}{}
	fast['['] = struct{}{}
	fast[']'] = struct{}{}
	fast['\n'] = struct{}{}
	fast['\t'] = struct{}{}
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		strings.FieldsFunc(s, func(c rune) bool { _, ok := fast[c]; return ok })
	}
}

func BenchmarkSplitFunc3(b *testing.B) {
	fast := make([]bool, 256)
	fast[' '] = true
	fast['('] = true
	fast[')'] = true
	fast['['] = true
	fast[']'] = true
	fast['\n'] = true
	fast['\t'] = true
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		strings.FieldsFunc(s, func(c rune) bool { return fast[c] })
	}
}

func BenchmarkSplitAny2(b *testing.B) {
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		splitAny2(s, " \n\t()[]")
	}
}

func BenchmarkSplitAny3(b *testing.B) {
	s := setupHelper(b)
	b.SetBytes(int64(len(s)))
	for i := 0; i < b.N; i++ {
		splitAny3(s, " \n\t()[]")
	}
}
