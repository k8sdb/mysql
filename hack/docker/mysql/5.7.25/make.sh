#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/mysql

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mysql
DB_VERSION=5.7.25
TAG="$DB_VERSION"

build() {
    pushd "$REPO_ROOT/hack/docker/mysql/$DB_VERSION"

    # Download Peer-finder
    # ref: peer-finder: https://github.com/kubernetes/contrib/tree/master/peer-finder
    # wget peer-finder: https://github.com/kubernetes/charts/blob/master/stable/mongodb-replicaset/install/Dockerfile#L18
    wget -qO peer-finder https://github.com/kmodules/peer-finder/releases/download/v1.0.1-ac/peer-finder
    chmod +x peer-finder

    local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
    echo $cmd; $cmd

    rm peer-finder
    popd
}

binary_repo $@
