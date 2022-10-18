package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/authhandler"
)

var twitterChannel chan [3]string

func twitterHandler(u string) (string, string, error) {
	err := browser.OpenURL(u)
	if err != nil {
		return "", "", err
	}
	s := <-twitterChannel
	fmt.Println(s)
	if s[2] != "" {
		return "", "", errors.New(s[2])
	}
	return s[0], s[1], nil
}

func main() {
	// TODO: Pixiv
	// TODO: Save keys in multiSav.json
	// TODO: Switch all personal APIs to use oauth2 and autorefresh
	hndlr := http.NewServeMux()
	hndlr.HandleFunc("/twitter", ServeTwitter)
	hndlr.Handle("/", http.NotFoundHandler())
	twitterChannel = make(chan [3]string)
	srv := &http.Server{Addr: ":5738", ReadHeaderTimeout: time.Second * 5, Handler: hndlr}
	go srv.ListenAndServe()
	config := &oauth2.Config{
		ClientID:     TwitterID,
		ClientSecret: TwitterSecret,
		RedirectURL:  "http://localhost:5738/twitter",
		Scopes:       []string{"tweet.read", "users.read", "list.read", "offline.access", "bookmark.read", "bookmark.write"},
	}
	config.Endpoint.AuthURL = "https://twitter.com/i/oauth2/authorize"
	config.Endpoint.TokenURL = "https://api.twitter.com/2/oauth2/token"
	var stateRand [32]byte
	rand.Read(stateRand[:])
	state := base64.NewEncoding("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZZZ").WithPadding(base64.NoPadding).EncodeToString(stateRand[:])
	var pkceRand [32]byte
	rand.Read(pkceRand[:])
	pkce := base64.NewEncoding("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZZZ").WithPadding(base64.NoPadding).EncodeToString(pkceRand[:])
	pkce2 := sha256.Sum256([]byte(pkce))
	pkce3 := base64.RawURLEncoding.EncodeToString(pkce2[:])
	// pkce3 := strings.TrimRight(strings.ReplaceAll(strings.ReplaceAll(string(pkce2[:]), "+", "-"), "/", "_"), "=")
	twitterToken := authhandler.TokenSourceWithPKCE(context.Background(), config, state, twitterHandler, &authhandler.PKCEParams{
		Verifier:        pkce,
		ChallengeMethod: "S256",
		Challenge:       pkce3,
	})
	fmt.Println(twitterToken.Token())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	srv.Shutdown(context.Background())
}

func ServeTwitter(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	err := req.ParseForm()
	if err != nil {
		s := fmt.Sprintf("failed to parse response data: %s\n", err.Error())
		fmt.Println(s)
		w.Header().Add("Content-Length", strconv.Itoa(len(s)))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(s))
		return
	}
	var s2 [3]string
	s2[0] = req.FormValue("code")
	s2[1] = req.FormValue("state")
	s2[2] = req.FormValue("error")
	twitterChannel <- s2
}
