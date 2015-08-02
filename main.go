package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/solarnz/starred-releases/feed"
	"golang.org/x/oauth2"
	githubOauth "golang.org/x/oauth2/github"
)

func main() {
	var httpBind string
	var clientID string
	var clientSecret string
	flag.StringVar(
		&httpBind, "http", os.Getenv("FEED_HTTP"),
		"The address to bind the server to",
	)
	flag.StringVar(
		&clientID, "client-id", os.Getenv("FEED_CLIENT_ID"),
		"The github oauth client ID of the application",
	)
	flag.StringVar(
		&clientSecret, "client-secret", os.Getenv("FEED_CLIENT_SECRET"),
		"The github oauth client secret of the application",
	)
	flag.Parse()

	if httpBind == "" {
		httpBind = ":8080"
	}

	var oauthConfig *oauth2.Config
	if clientID != "" && clientSecret != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{},
			Endpoint:     githubOauth.Endpoint,
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		if oauthConfig == nil {
			w.WriteHeader(500)
			fmt.Println("You must specify client-id and client-secret to generate access tokens.")
			fmt.Println(
				"Alternatively you can generate a personal access token and go to " +
					r.Host + "/<username>/<token>/atom.xml",
			)
			return
		}
		if r.FormValue("code") == "" {
			url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
			http.Redirect(w, r, url, 302)
		} else {
			code := r.FormValue("code")
			t, err := oauthConfig.Exchange(oauth2.NoContext, code)
			if err != nil {
				w.WriteHeader(500)
				fmt.Println(err)
				return
			}
			gh := github.NewClient(oauthConfig.Client(oauth2.NoContext, t))
			u, _, err := gh.Users.Get("")
			if err != nil {
				w.WriteHeader(500)
				fmt.Println(err)
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/%s/%s/atom.xml", *u.Login, t.AccessToken), 302)
		}
	})
	r.HandleFunc("/{user}/{token}/atom.xml", func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: v["token"]})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		c := feed.NewFeed(tc)

		feed, err := c.BuildFeed(r.URL.RequestURI(), v["user"])
		if err != nil {
			w.WriteHeader(500)
			fmt.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		w.Write(feed)
	})

	http.ListenAndServe(httpBind, r)
}
