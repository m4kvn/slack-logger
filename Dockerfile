FROM alpine
MAINTAINER masahirosaito
RUN apk --update add go git libc-dev tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*
COPY main.go /root/go/src/slack-logger/main.go
RUN go get -t slack-logger && \
    go install slack-logger && \
    mkdir app
WORKDIR app
ENTRYPOINT /root/go/bin/slack-logger -token $TOKEN -channel $CHANNEL -time $TIME