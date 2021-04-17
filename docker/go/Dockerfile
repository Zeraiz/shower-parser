FROM golang:1.16.0-alpine3.13

RUN apk add --no-cache git

ENV MUSL_LOCPATH=/usr/local/share/i18n/locales/musl
RUN apk add --update git cmake make musl-dev gcc gettext-dev libintl
RUN cd /tmp && git clone https://github.com/rilian-la-te/musl-locales.git
RUN cd /tmp/musl-locales && cmake . && make && make install

ENV LANG=ru_RU.UTF-8 \
    LANGUAGE=ru_RU.UTF-8

WORKDIR /go/src/app

COPY ./docker/go/docker-entrypoint.sh /

ENV GO111MODULE=auto
ENV APP_HOME /go/src/app
ENV GOPATH /go/src/app
ENV CGO_ENABLED=0

ENTRYPOINT ["sh", "/docker-entrypoint.sh"]