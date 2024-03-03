#!/bin/bash


GOOS=linux GOARCH=arm GOARM=5 go build  -o bin/pi-brew main.go
scp bin/pi-brew pi@tiltpi.local:
ssh pi@tiltpi.local
