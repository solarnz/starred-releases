# starred-repositories

A simple go server that will fetch your starred repositories, fetch the
releases for those repositories and generate an atom feed of those releases.

## Configuration

You will need to generate a github personal access token. You can do this by
going to the settings page on github, going to the "Personal access tokens"
section and clicking on "Generate new token"

## Running

### Docker

    docker build . -t starred-repositories
    docker run -e FEED_USER=YOURUSERNAME -e FEED_TOKEN=PERSONALACCESSTOKEN starred-repositories

### Locally

    go get github.com/solarnz/starred-repositories
