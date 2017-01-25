#!/bin/sh 

GOOS=linux GOARCH=amd64 go build -o example example.go 
docker build -t mbexample .