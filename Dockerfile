FROM golang:alpine AS build
MAINTAINER parker714@foxmail.com

WORKDIR /apps
ADD . /apps
RUN cd /apps/cmd && go build -o cron

FROM alpine
WORKDIR /apps
COPY --from=build /apps/cmd/cron /
EXPOSE 7570
CMD ["/cron"]