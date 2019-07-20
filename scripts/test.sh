#!/usr/bin/env bash

GOMAXPROCS=1 go test -timeout 90s $(go list ./...)
GOMAXPROCS=4 go test -timeout 90s -race $(go list ./...)
