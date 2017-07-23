FROM golang:onbuild
MAINTAINER m411momo (MasahiroSaito)
RUN go get github.com/mattn/go-sqlite3
ENTRYPOINT go run slacklogger.go $TOKEN