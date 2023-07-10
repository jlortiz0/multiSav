package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/jlortiz0/multisav/pixivapi"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/authhandler"
)

var respChannel chan [3]string

func twitterHandler(u string) (string, string, error) {
	err := browser.OpenURL(u)
	if err != nil {
		return "", "", err
	}
	s := <-respChannel
	if s[2] != "" {
		return "", "", errors.New(s[2])
	}
	return s[0], s[1], nil
}

func redditHandler(u string) (string, string, error) {
	err := browser.OpenURL(u + "&duration=permanent")
	if err != nil {
		return "", "", err
	}
	s := <-respChannel
	if s[2] != "" {
		return "", "", errors.New(s[2])
	}
	return s[0], s[1], nil
}

func pixivHandler(u string) (string, string, error) {
	err := browser.OpenURL(u)
	if err != nil {
		return "", "", err
	}
	var s string
	fmt.Scanf("https://app-api.pixiv.net/web/v1/users/auth/pixiv/callback?%s\n", &s)
	s3, _ := url.ParseQuery(s)
	return s3.Get("code"), s3.Get("state"), nil
}

func waitToDie(ch chan os.Signal, srv *http.Server) {
	signal.Notify(ch, os.Interrupt)
	<-ch
	srv.Shutdown(context.Background())
	fmt.Println("Goodbye...")
	os.Exit(0)
}

func main() {
	if _, err := os.Stat("../superAuthorizer"); err == nil {
		os.Chdir("..")
	}
	if _, err := os.Stat("jlortiz_TEST"); err == nil {
		os.Chdir("jlortiz_TEST")
	}
	loadSaveData()
	hndlr := http.NewServeMux()
	hndlr.HandleFunc("/twitter", ServeTwitter)
	hndlr.HandleFunc("/reddit", ServeTwitter)
	hndlr.Handle("/", http.NotFoundHandler())
	respChannel = make(chan [3]string)
	srv := &http.Server{Addr: ":5738", ReadHeaderTimeout: time.Second * 5, Handler: hndlr}
	go srv.ListenAndServe()
	ch := make(chan os.Signal, 1)
	go waitToDie(ch, srv)

	var stateRand [32]byte
	rand.Read(stateRand[:])
	state := base64.NewEncoding("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZZZ").WithPadding(base64.NoPadding).EncodeToString(stateRand[:])
	config := &oauth2.Config{
		ClientID:     TwitterID,
		ClientSecret: TwitterSecret,
		RedirectURL:  "http://localhost:5738/twitter",
		Scopes:       []string{"tweet.read", "users.read", "list.read", "offline.access", "bookmark.read", "bookmark.write"},
	}
	config.Endpoint.AuthURL = "https://twitter.com/i/oauth2/authorize"
	config.Endpoint.TokenURL = "https://api.twitter.com/2/oauth2/token"
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

	config = &oauth2.Config{
		ClientID:     RedditID,
		ClientSecret: RedditSecret,
		RedirectURL:  "http://localhost:5738/reddit",
		Scopes:       []string{"history", "identity", "read", "save", "subscribe"},
	}
	config.Endpoint.AuthURL = "https://www.reddit.com/api/v1/authorize"
	config.Endpoint.TokenURL = "https://www.reddit.com/api/v1/access_token"
	redditToken := authhandler.TokenSource(context.Background(), config, state, redditHandler)

	config = &oauth2.Config{
		ClientID:     pixivapi.Client_ID,
		ClientSecret: pixivapi.Client_secret,
		RedirectURL:  "https://app-api.pixiv.net/web/v1/users/auth/pixiv/callback",
		Scopes:       []string{""},
	}
	config.Endpoint.AuthURL = "https://oauth.secure.pixiv.net/auth/authorize"
	config.Endpoint.TokenURL = "https://oauth.secure.pixiv.net/auth/token"
	pixivToken := authhandler.TokenSourceWithPKCE(context.Background(), config, state, pixivHandler, &authhandler.PKCEParams{
		Verifier:        pkce,
		ChallengeMethod: "S256",
		Challenge:       pkce3,
	})

	for {
		fmt.Print("Super Authorizer\n1. Twitter\n2. Pixiv\n3. Reddit\n4. Lemmy\n5. Exit\n\nSel: ")
		i := 5
		n, _ := fmt.Scanf("%d\n", &i)
		if n != 1 {
			i = 5
		}
		switch i {
		case 5:
			ch <- nil
			(chan int)(nil) <- 0
		case 1:
			fmt.Println("Authorize the app in your browser")
			t, err := twitterToken.Token()
			if err != nil {
				fmt.Printf("An error occured: %s\n", err.Error())
			} else {
				saveData["Twitter"] = t.RefreshToken
				saveSaveData()
				fmt.Println("Success! Token has been saved.")
			}
			fmt.Println("Press enter to continue.")
			fmt.Scanf("\n")
		case 2:
			fmt.Println("When the page opens in your browser, sign in. You will see a message saying \"Invalid request\"\nPaste the URL of the page where you get that message below:")
			t, err := pixivToken.Token()
			if err != nil {
				fmt.Printf("An error occured: %s\n", err.Error())
			} else {
				saveData["Pixiv"] = t.RefreshToken
				saveSaveData()
				fmt.Println("Success! Token has been saved.")
			}
			fmt.Println("Press enter to continue.")
			fmt.Scanf("\n")
		case 3:
			fmt.Println("Authorize the app in your browser")
			t, err := redditToken.Token()
			if err != nil {
				fmt.Printf("An error occured: %s\n", err.Error())
			} else {
				saveData["Reddit"] = t.RefreshToken
				saveSaveData()
				fmt.Println("Success! Token has been saved.")
			}
			fmt.Println("Press enter to continue.")
			fmt.Scanf("\n")
		case 4:
			fmt.Print("Lemmy site: ")
			var site, user, pass string
			fmt.Scanf("%s\n", &site)
			if strings.ContainsRune(site, '/') {
				fmt.Println("Only put the actual site name, no http or slashes")
			} else {
				fmt.Print("Username or email (leave blank for anonymous): ")
				fmt.Scanf("%s\n", &user)
				if user != "" {
					fmt.Print("Password: ")
					fmt.Scanf("%s\n", &pass)
				}
				saveData["Lemmy"] = struct{ Base, User, Pass string }{
					site, user, pass,
				}
				saveSaveData()
				fmt.Println("Hopefully this works...")
			}
			fmt.Println("Press any key to continue.")
			fmt.Scanf("\n")
		default:
			fmt.Println("Invalid selection.")
		}
	}
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
	respChannel <- s2
	w.Write([]byte("success check console for more details"))
}
