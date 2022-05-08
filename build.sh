#!/bin/bash
GOOS=windows GOARCH=amd64 go build -o img_grab_win.exe run.go
GOOS=linux GOARCH=amd64 go build -o img_grab_linux run.go
