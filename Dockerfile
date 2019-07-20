FROM golang:1.12 AS build
MAINTAINER parker714@foxmail.com

WORKDIR /apps
ADD . /apps
RUN cd /apps/cmd && go build -mod=vendor -o cron

FROM alpine
WORKDIR /apps
COPY --from=build /apps/cmd/cron /
EXPOSE 7570
CMD ["/cron"]