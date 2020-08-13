#!/bin/bash

export GOARCH=arm
export GOARM=7
export GOOS=linux
export GOPATH=/Users/chren/entwicklung/go

serverAddress=192.168.2.43
projectName=orion
moduleName=orion.misc
now=$(date +"%m%d%Y-%H%M%S")

#echo 'Dir: /home/chrisu/${projectName}/${moduleName}/${moduleName}'

go build -o $moduleName main.go
ssh chrisu@$serverAddress "sudo systemctl stop ${moduleName}"
ssh chrisu@$serverAddress "cp /home/chrisu/${projectName}/${moduleName}/${moduleName} /home/chrisu/${projectName}/${moduleName}/${moduleName}_BAK_${now}"
scp $moduleName chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName
scp /Users/chren/entwicklung/go/src/$moduleName/resources/config_prod.toml chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName/config.toml
ssh chrisu@$serverAddress "sudo systemctl start ${moduleName}"