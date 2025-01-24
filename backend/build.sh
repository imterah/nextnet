#!/usr/bin/env bash
pushd sshbackend > /dev/null
echo "building sshbackend"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

pushd dummybackend > /dev/null
echo "building dummybackend"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

pushd externalbackendlauncher > /dev/null
echo "building externalbackendlauncher"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null

pushd sshappbackend/remote-code > /dev/null
echo "building sshappbackend/remote-code"
if [ ! -d bin ]; then
    mkdir bin
fi

echo " - building for arm64"
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o bin/rt-arm64 .
echo " - building for arm"
GOOS=linux GOARCH=arm go build -ldflags="-s -w" -trimpath -o bin/rt-arm .
echo " - building for amd64"
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/rt-amd64 .
echo " - building for i386"
GOOS=linux GOARCH=386 go build -ldflags="-s -w" -trimpath -o bin/rt-386 .
popd > /dev/null

pushd sshappbackend/local-code > /dev/null
echo "building sshappbackend/local-code"
go build -ldflags="-s -w" -trimpath -o sshappbackend .
popd > /dev/null

pushd api > /dev/null
echo "building api"
go build -ldflags="-s -w" -trimpath .
popd > /dev/null
