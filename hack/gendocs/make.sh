#!/usr/bin/env bash

pushd $GOPATH/src/github.com/k8sdb/mysql/hack/gendocs
go run main.go
popd
