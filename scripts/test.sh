#!/usr/bin/env bash
set -e

GOMAXPROCS=1 go test -timeout 90s $(go list ../...)
GOMAXPROCS=4 go test -timeout 90s -race $(go list ../...)

# no tests, but a build is something
# go build -o ../crond ../cmd
