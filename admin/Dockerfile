FROM golang

WORKDIR /go/src/github.com/onlineconf/onlineconf/admin/go

COPY go/go.* ./
RUN go mod download

COPY go .
RUN go build -o onlineconf-admin


FROM node

WORKDIR /usr/src/onlineconf-admin

COPY js/package*.json ./
RUN npm install

COPY js .
RUN npm run build


FROM debian

COPY --from=0 /go/src/github.com/onlineconf/onlineconf/admin/go/onlineconf-admin /usr/local/bin/onlineconf-admin
COPY --from=1 /usr/src/onlineconf-admin/build /var/www/onlineconf-admin
COPY static /var/www/onlineconf-admin/classic
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 80/tcp
EXPOSE 443/tcp
