#!/bin/bash


GOOS=linux GOARCH=arm GOARM=5 go build  -o scanner main.go
scp scanner pi@tiltpi.local:
ssh pi@tiltpi.local
