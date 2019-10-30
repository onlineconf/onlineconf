FROM golang

WORKDIR /go/src/github.com/onlineconf/onlineconf/updater

COPY go.* ./
RUN go mod download

COPY updater ./updater
COPY *.go ./
RUN go build -o onlineconf-updater


FROM debian

COPY --from=0 /go/src/github.com/onlineconf/onlineconf/updater/onlineconf-updater /usr/local/bin/onlineconf-updater
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
