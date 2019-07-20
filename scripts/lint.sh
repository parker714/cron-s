#!/usr/bin/env bash

golint $(go list ./...)

golangci-lint run