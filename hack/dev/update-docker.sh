#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/mysql

# $REPO_ROOT/hack/docker/mysql/5.7/make.sh
# $REPO_ROOT/hack/docker/mysql/8.0/make.sh

$REPO_ROOT/hack/docker/mysql-tools/5.7/make.sh build
$REPO_ROOT/hack/docker/mysql-tools/5.7/make.sh push

$REPO_ROOT/hack/docker/mysql-tools/8.0/make.sh build
$REPO_ROOT/hack/docker/mysql-tools/8.0/make.sh push


# $REPO_ROOT/hack/docker/my-operator/make.sh build
# $REPO_ROOT/hack/docker/my-operator/make.sh push

