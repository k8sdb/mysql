#!/usr/bin/env bash

pushd $GOPATH/src/kubedb.dev/mysql/hack/gendocs
go run main.go
popd
