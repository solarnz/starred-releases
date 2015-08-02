FROM golang
ADD . /go/src/github.com/solarnz/starred-releases/
RUN go get github.com/solarnz/starred-releases
RUN go install github.com/solarnz/starred-releases

ENTRYPOINT /go/bin/starred-releases
ENV FEED_USER ""
ENV FEED_TOKEN ""
ENV FEED_HTTP ":80"
EXPOSE 80
