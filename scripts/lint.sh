#!/usr/bin/env bash

golint $(go list ../...)

cd ../ && golangci-lint run