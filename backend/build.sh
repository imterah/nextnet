#!/usr/bin/env bash
pushd sshbackend
go build .
strip sshbackend
popd

pushd dummybackend
go build .
strip dummybackend
popd

pushd externalbackendlauncher
go build .
strip externalbackendlauncher
popd

pushd api
go build .
strip api
popd
