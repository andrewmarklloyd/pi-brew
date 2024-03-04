#!/bin/bash


GOOS=linux GOARCH=arm GOARM=5 go build  -o bin/pi-brew main.go
export DD_APP_KEY=$(op read ${OP_DD_API_KEY_PATH})
export DD_API_KEY=$(op read ${OP_DD_APP_KEY_PATH})
ansible-playbook ansible/playbook.yaml -K --inventory-file ansible/inventory.yaml
