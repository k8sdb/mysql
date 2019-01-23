#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}

IMG=mysql

DB_VERSION=8.0.3
TAG="$DB_VERSION"

docker pull $IMG:$DB_VERSION

docker tag $IMG:$DB_VERSION "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"
