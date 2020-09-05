#!/bin/bash

# Copyright AppsCode Inc. and Contributors
#
# Licensed under the AppsCode Community License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.




set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/kubedb.dev/mysql

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=mysql
DB_VERSION=5.7.25
TAG="$DB_VERSION"

build() {
  pushd "$REPO_ROOT/hack/docker/mysql/$DB_VERSION"

  # Download Peer-finder
  # ref: peer-finder: https://github.com/kmodules/peer-finder/releases/download/v1.0.1-ac/peer-finder
  # wget peer-finder: https://github.com/kubernetes/charts/blob/master/stable/mongodb-replicaset/install/Dockerfile#L18
  wget -qO peer-finder https://github.com/kmodules/peer-finder/releases/download/v1.0.1-ac/peer-finder
  chmod +x peer-finder

  local cmd="docker build --pull -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd
  $cmd

  rm peer-finder
  popd
}

binary_repo $@
