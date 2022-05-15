#!/bin/bash
set -euo pipefail

rm -rf ./img_grab_*
GOOS=linux GOARCH=amd64 go build -ldflags=-w -o img_grab_linux run.go
GOOS=windows GOARCH=amd64 go build -ldflags=-w -o img_grab_win.exe run.go
