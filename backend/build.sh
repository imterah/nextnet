#!/usr/bin/env bash
pushd sshbackend
GOOS=linux go build .
strip sshbackend
popd

pushd dummybackend
GOOS=linux go build .
strip dummybackend
popd

pushd externalbackendlauncher
go build .
strip externalbackendlauncher
popd

pushd api
GOOS=linux go build .
strip api
popd
