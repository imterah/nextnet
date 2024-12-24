#!/usr/bin/env bash
pushd sshbackend
CGO_ENABLED=0 GOOS=linux go build .
strip sshbackend
popd

pushd dummybackend
CGO_ENABLED=0 GOOS=linux go build .
strip dummybackend
popd

pushd externalbackendlauncher
go build .
strip externalbackendlauncher
popd

pushd api
CGO_ENABLED=0 GOOS=linux go build .
strip api
popd
