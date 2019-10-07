FROM golang

WORKDIR /go/src/github.com/onlineconf/onlineconf/updater/go

COPY go/go.* ./
RUN go mod download

COPY go .
RUN go build -o onlineconf-updater


FROM debian

COPY --from=0 /go/src/github.com/onlineconf/onlineconf/updater/go/onlineconf-updater /usr/local/bin/onlineconf-updater
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]
