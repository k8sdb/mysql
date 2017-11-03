#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/k8sdb/mysql

source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

IMG=mysql
TAG=8.0

build() {
    pushd $REPO_ROOT/hack/docker/mysql/8.0
    docker build -t kubedb/$IMG:$TAG .
	popd
}

docker_push() {
        docker push kubedb/$IMG:$TAG
}

docker_release() {
        docker push kubedb/$IMG:$TAG
}

docker_check() {
        echo "Chcking $IMG ..."
        name=$(date +%s | sha256sum | base64 | head -c 8 ; echo)
        docker run -d -P -it --name=$name kubedb/$IMG:$TAG
        docker exec -it $name ps aux
        sleep 5
        docker exec -it $name ps aux
        docker stop $name && docker rm $name
}

binary_repo $@
