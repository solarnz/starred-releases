package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/solarnz/starred-releases/feed"
	"golang.org/x/oauth2"
)

func main() {
	var user string
	var personalAccessToken string
	var httpBind string
	flag.StringVar(&user, "user", os.Getenv("FEED_USER"), "The username to fetch the feed for")
	flag.StringVar(&personalAccessToken, "access-token", os.Getenv("FEED_TOKEN"), "Your personal access token for github")
	flag.StringVar(&httpBind, "http", os.Getenv("FEED_HTTP"), "The address to bind the server to")
	flag.Parse()

	if user == "" {
		log.Fatal("user or FEED_USER must be specified")
	}
	if personalAccessToken == "" {
		log.Fatal("accessToken or FEED_TOKEN must be specified")
	}
	if httpBind == "" {
		httpBind = ":8080"
	}

	http.HandleFunc("/feed", func(w http.ResponseWriter, r *http.Request) {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: personalAccessToken})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		c := feed.NewFeed(tc)

		feed, err := c.BuildFeed("http://"+r.Host+"/feed", user)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		w.Write(feed)
	})

	http.ListenAndServe(httpBind, nil)
}
