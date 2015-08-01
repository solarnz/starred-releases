package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"golang.org/x/tools/blog/atom"
)

const PAGE_SIZE = 100

type starredClient struct {
	github *github.Client
}

type Release struct {
	owner      string
	repository string
	github.RepositoryRelease
}

type ByDate []Release

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d ByDate) Less(i, j int) bool {
	return d[i].PublishedAt.Sub(d[j].PublishedAt.Time).Seconds() < 0
}

func (c starredClient) GetStarredRepositories(user string) ([]github.StarredRepository, error) {
	var repos []github.StarredRepository

	p := 0
	more := true
	a := c.github.Activity
	for more {
		o := github.ActivityListStarredOptions{
			ListOptions: github.ListOptions{
				Page:    p,
				PerPage: PAGE_SIZE,
			},
		}
		s, _, err := a.ListStarred(user, &o)
		if err != nil {
			return nil, err
		}

		for _, r := range s {
			repos = append(repos, r)
		}

		if len(s) < PAGE_SIZE {
			more = false
		}
	}

	return repos, nil
}

func (c starredClient) GetRepositoryReleases(user, repo string) ([]github.RepositoryRelease, error) {
	var releases []github.RepositoryRelease

	p := 0
	more := true
	rs := c.github.Repositories
	for more {
		o := github.ListOptions{
			Page:    p,
			PerPage: 100,
		}

		r, _, err := rs.ListReleases(user, repo, &o)
		if err != nil {
			return nil, err
		}

		for _, r := range r {
			releases = append(releases, r)
		}

		if len(r) < PAGE_SIZE {
			more = false
		}
	}

	return releases, nil
}

func (c starredClient) BuildFeed(feedID, user string) ([]byte, error) {
	repos, err := c.GetStarredRepositories(user)
	if err != nil {
		return nil, err
	}

	var releases []Release
	var errors []error
	releasesLock := sync.Mutex{}
	errorsLock := sync.Mutex{}
	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(repo github.StarredRepository) {
			owner := *(repo.Repository.Owner.Login)
			name := *(repo.Repository.Name)
			repositoryReleases, err := c.GetRepositoryReleases(owner, name)
			if err != nil {
				errorsLock.Lock()
				errors = append(errors, err)
				errorsLock.Unlock()
				wg.Done()
				return
			}

			for _, release := range repositoryReleases {
				releasesLock.Lock()
				releases = append(releases, Release{
					owner:             owner,
					repository:        name,
					RepositoryRelease: release,
				})
				releasesLock.Unlock()
			}
			wg.Done()
		}(repo)
	}

	wg.Wait()
	if len(errors) != 0 {
		return nil, fmt.Errorf("encountered the following errors: %s", errors)
	}

	sort.Sort(sort.Reverse(ByDate(releases)))

	var entries []*atom.Entry
	for _, release := range releases {
		entries = append(entries, &atom.Entry{
			Title: fmt.Sprintf(
				"[%s/%s] %s (%s)",
				release.owner, release.repository, *release.Name, *release.TagName,
			),
			ID:        feedID + "/" + strconv.Itoa(*release.ID),
			Updated:   atom.Time(release.PublishedAt.Time),
			Published: atom.Time(release.PublishedAt.Time),
			Author: &atom.Person{
				Name: release.owner,
			},
			Link: []atom.Link{{
				Rel:  "alternate",
				Href: *release.HTMLURL,
			}},
			Content: &atom.Text{
				Type: "html",
				Body: fmt.Sprintf("<pre>%s</pre>", html.EscapeString(*release.Body)),
			},
		})
	}

	feed := atom.Feed{
		Title:   fmt.Sprintf("Starred Github Releases for %s", user),
		ID:      feedID,
		Updated: atom.Time(time.Now()),
		Entry:   entries,
		Link: []atom.Link{{
			Rel:  "self",
			Href: feedID,
		}},
	}
	b, err := xml.Marshal(&feed)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func main() {
	var user string
	var personalAccessToken string
	var httpBind string
	flag.StringVar(&user, "user", os.Getenv("FEED_USER"), "The username to fetch the feed for")
	flag.StringVar(&personalAccessToken, "accessToken", os.Getenv("FEED_TOKEN"), "Your personal access token for github")
	flag.StringVar(&httpBind, "http", os.Getenv("FEED_HTTP"), "The address to bind the server to")

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
		c := starredClient{
			github: github.NewClient(tc),
		}

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
