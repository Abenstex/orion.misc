set GOARCH=arm
set GOARM=7
set GOOS=linux
set GOPATH=C:\Users\chris\go

set serverAddress=192.168.2.42
set moduleName=orion.misc
set rootPath=/home/chrisu/orion/releases
set releaseVersion=0.2.1
set osArch=arm

echo 'Building ' %moduleName%
go build -o %moduleName% main.go