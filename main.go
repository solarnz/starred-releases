package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/solarnz/starred-releases/feed"
	"golang.org/x/oauth2"
)

func main() {
	var httpBind string
	flag.StringVar(&httpBind, "http", os.Getenv("FEED_HTTP"), "The address to bind the server to")
	flag.Parse()

	if httpBind == "" {
		httpBind = ":8080"
	}

	r := mux.NewRouter()
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
