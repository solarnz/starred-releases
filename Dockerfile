FROM golang
ADD . /go/src/github.com/solarnz/starred-releases/
RUN go get github.com/solarnz/starred-releases
RUN go install github.com/solarnz/starred-releases

ENTRYPOINT /go/bin/starred-releases
ENV FEED_USER ""
ENV FEED_TOKEN ""
EXPOSE 8080
