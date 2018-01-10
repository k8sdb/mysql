#!/bin/bash
set -xeou pipefail

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mysql
TAG=8.0

docker pull $IMG:$TAG

docker tag $IMG:$TAG "$DOCKER_REGISTRY/$IMG:$TAG"
docker push "$DOCKER_REGISTRY/$IMG:$TAG"

docker tag $IMG:$TAG "$DOCKER_REGISTRY/$IMG:8"
docker push "$DOCKER_REGISTRY/$IMG:8"
