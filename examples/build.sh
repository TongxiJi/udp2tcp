#!/usr/bin/env bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o udp2tcp