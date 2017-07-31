FROM alpine
MAINTAINER masahirosaito
RUN apk --update add go git libc-dev tzdata
RUN cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/* && \
    mkdir -p /go/src/app
ENV GOPATH=/go
WORKDIR /go/src/app
ENTRYPOINT go get -t && go build && ./app