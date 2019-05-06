FROM golang:latest AS build

RUN mkdir -p /go/src/github.com/parker714/cron-s
COPY . /go/src/github.com/parker714/cron-s

WORKDIR /go/src/github.com/parker714/cron-s

RUN wget -O /bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 \
 && chmod +x /bin/dep \
 && /bin/dep ensure \
 && ./test.sh CGO_ENABLED=0 make DESTDIR=/opt PREFIX=/cron BLDFLAGS='-ldflags="-s -w"' install

FROM alpine:3.7

EXPOSE 8570 7570

COPY --from=build /opt/cron/bin/ /usr/local/bin/
CMD ["apps/crond/crond"]