package pixivapi_test

import (
	"encoding/json"
	"os"
	"testing"

	"jlortiz.org/redisav/pixivapi"
)

func TestLogin(T *testing.T) {
	loginHelper(T)
}

func loginHelper(T *testing.T) *pixivapi.Client {
	T.Helper()
	data := make([]byte, 1024)
	f, err := os.Open("../redditapi/login.json")
	if err != nil {
		T.Fatalf("Failed to open login data file: %s", err.Error())
	}
	n, err := f.Read(data)
	if err != nil {
		T.Fatalf("Failed to read login data: %s", err.Error())
	}
	f.Close()
	var fields struct {
		PixivId     string
		PixivSecret string
		PixivToken  string
		PixivAccess string
	}
	err = json.Unmarshal(data[:n], &fields)
	if err != nil {
		T.Fatalf("Failed to decode login data: %s", err.Error())
	}
	red := pixivapi.NewClient(fields.PixivId, fields.PixivSecret)
	if fields.PixivAccess != "" {
		red.SetToken(fields.PixivAccess, fields.PixivToken)
	} else {
		err = red.Login(fields.PixivToken)
		if err != nil {
			T.Fatalf("Failed to log in: %s", err.Error())
		}
		T.Log("New refresh token: " + red.RefreshToken())
	}
	return red
}

func TestIllust(T *testing.T) {
	p := loginHelper(T)
	ret, err := p.GetIllust(101469224)
	if err != nil {
		T.Fatal(err)
	}
	T.Log(ret)
}
