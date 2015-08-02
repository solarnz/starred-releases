package feed

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.com/microcosm-cc/bluemonday"
	md "github.com/shurcooL/github_flavored_markdown"
	"golang.org/x/tools/blog/atom"
)

const PAGE_SIZE = 100

type Feed struct {
	github *github.Client
}

type Release struct {
	owner      string
	repository string
	github.RepositoryRelease
}

func (r Release) Releaser() string {
	if len(r.Assets) > 1 {
		return *r.Assets[0].Uploader.Login
	}
	return r.owner
}

func (r Release) SanitisedBody() string {
	var i string
	if r.Body == nil {
		i = ""
	} else {
		i = *r.Body
	}
	s := bluemonday.UGCPolicy()
	b := md.Markdown([]byte(i))
	return string(s.SanitizeBytes(b))
}

type ByDate []Release

func (d ByDate) Len() int      { return len(d) }
func (d ByDate) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d ByDate) Less(i, j int) bool {
	return d[i].PublishedAt.Sub(d[j].PublishedAt.Time).Seconds() < 0
}

func (c Feed) GetStarredRepositories(user string) ([]github.StarredRepository, error) {
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

func (c Feed) GetRepositoryReleases(user, repo string) ([]github.RepositoryRelease, error) {
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

func (c Feed) BuildFeed(feedID, user string) ([]byte, error) {
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
				Name: release.Releaser(),
			},
			Link: []atom.Link{{
				Rel:  "alternate",
				Href: *release.HTMLURL,
			}},
			Content: &atom.Text{
				Type: "html",
				Body: release.SanitisedBody(),
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

func NewFeed(h *http.Client) Feed {
	if h == nil {
		h = http.DefaultClient
	}

	return Feed{
		github: github.NewClient(h),
	}
}
