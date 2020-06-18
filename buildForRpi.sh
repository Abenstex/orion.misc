#!/bin/bash

export GOARCH=arm
export GOARM=7
export GOOS=linux
export GOPATH=/Users/chren/entwicklung/go

serverAddress=192.168.2.43
projectName=orion
moduleName=orion.misc

go build -o $moduleName main.go
scp $moduleName chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName
scp /Users/chren/entwicklung/go/src/$moduleName/resources/config_prod.toml chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName/config.toml
