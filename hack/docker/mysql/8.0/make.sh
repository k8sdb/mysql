#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mysql
TAG=8.0
ALT_TAG=8
PATCH=8.0.3

docker pull $IMG:$PATCH

docker tag $IMG:$PATCH "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"

docker tag $IMG:$PATCH "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
docker push "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
