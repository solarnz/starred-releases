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
    docker run -p "8080:80" starred-releases

### Locally

    go get github.com/solarnz/starred-releases
    starred-releases

## Usage

You can then access the atom feed at `http://localhost:8080/<github username>/<personal access token>/atom.xml`

Alternatively, you can use the public instance. Generate your feed by going to http://feedr.sloo.co/generate
