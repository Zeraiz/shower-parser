FROM golang:1.16.0-alpine3.13

RUN apk add --no-cache git

WORKDIR /go/src/app

COPY ./docker/go/docker-entrypoint.sh /

ENV GO111MODULE=auto
ENV APP_HOME /go/src/app
ENV GOPATH /go/src/app
ENV CGO_ENABLED=0

ENTRYPOINT ["sh", "/docker-entrypoint.sh"]