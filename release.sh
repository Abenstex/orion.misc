#!/bin/bash

export GOARCH=arm
export GOARM=7
export GOOS=linux
export GOPATH=/Users/chren/entwicklung/go

serverAddress=192.168.2.42
moduleName=orion.misc
rootPath=/home/chrisu/orion/releases
releaseVersion=0.0.1
osArch=arm

#echo 'Dir: /home/chrisu/${projectName}/${moduleName}/${moduleName}'

echo 'Building ${moduleName}'
go build -o $moduleName main.go
#ssh chrisu@$serverAddress "sudo systemctl stop ${moduleName}"
#ssh chrisu@$serverAddress "cp /home/chrisu/${projectName}/${moduleName}/${moduleName} /home/chrisu/${projectName}/${moduleName}/${moduleName}_BAK_${now}"
#scp $moduleName chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName
#scp /Users/chren/entwicklung/go/src/$moduleName/resources/config_prod.toml chrisu@$serverAddress:/home/chrisu/$projectName/$moduleName/config.toml
#ssh chrisu@$serverAddress "sudo systemctl start ${moduleName}"

echo 'Starting upload'
sftp -oPort=22 chrisu@$serverAddress <<EOF
mkdir ${rootPath}/${moduleName}/${osArch}/${releaseVersion}
cd ${rootPath}/${moduleName}/${osArch}/${releaseVersion}
put -pr $moduleName
exit
EOF
echo 'Finished upload'
