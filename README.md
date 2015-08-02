# starred-releases

A simple go server that will fetch your starred repositories, fetch the
releases for those repositories and generate an atom feed of those releases.

## Configuration

You will need to generate a github personal access token. You can do this by
going to the settings page on github, going to the "Personal access tokens"
section and clicking on "Generate new token"

## Running

### Docker

    docker build . -t starred-releases
    docker run -e FEED_USER=YOURUSERNAME -e FEED_TOKEN=PERSONALACCESSTOKEN -p "8080:80" starred-releases

### Locally

    go get github.com/solarnz/starred-releases
    starred-releases -user YOURUSERNAME -access-token ACCESSTOKEN

## Usage

You can then access the atom feed on http://localhost:8080/feed
