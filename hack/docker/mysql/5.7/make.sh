#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mysql
TAG=5.7
ALT_TAG=5

docker pull $IMG:$TAG

docker tag $IMG:$TAG "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"

docker tag $IMG:$TAG "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
docker push "$DOCKER_REGISTRY/$IMG:$ALT_TAG"
