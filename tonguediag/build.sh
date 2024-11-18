#!/bin/bash

env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build
systemctl stop tongue && cp tonguediag /usr/local/tonguediag/tonguediag && systemctl start tongue

