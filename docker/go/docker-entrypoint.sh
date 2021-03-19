#!/bin/sh
cd /go/src/app
go build -o ./builds main.go
vi /go/src/app/main.go