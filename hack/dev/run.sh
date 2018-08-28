#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedbci}
export KUBEDB_UNINSTALL=0
export KUBEDB_PURGE=0
export MINIKUBE=1
export SELF_HOSTED=0

REPO_ROOT="$GOPATH/src/github.com/kubedb/mysql"

show_help() {
  echo "run.sh - run kubedb operator"
  echo " "
  echo "run.sh [options]"
  echo " "
  echo "options:"
  echo "-h, --help                         show brief help"
  echo "    --uninstall                    uninstall kubedb deployment and staffs before running oprator"
  echo "    --purge                        purges kubedb crd objects and crds before running operator"
  echo "    --minikube                     run operator in local and connect with minikube"
  echo "    --selfhosted                  deploy operator in cluster"
}

while test $# -gt 0; do
  case "$1" in
    -h | --help)
      show_help
      exit 0
      ;;
    --minikube)
      export MINIKUBE=1
      export SELF_HOSTED=0
      shift
      ;;
    --selfhosted)
      export MINIKUBE=0
      export SELF_HOSTED=1
      shift
      ;;
    --uninstall)
      export KUBEDB_UNINSTALL=1
      shift
      ;;
    --purge)
      export KUBEDB_PURGE=1
      shift
      ;;
    *)
      show_help
      exit 1
      ;;
  esac
done

if [ "$KUBEDB_UNINSTALL" -eq 1 ]; then
  # delete webhooks and apiservices

  if [ "$KUBEDB_PURGE" -eq 1 ]; then
    $REPO_ROOT/hack/dev/toolbox.sh --uninstall --purge
  else
    $REPO_ROOT/hack/dev/toolbox.sh --uninstall
  fi
fi

# run in minikube
# cat $REPO_ROOT/hack/dev/mysql-catalog.yaml | $ONESSL envsubst | kubectl apply -f -

if [ "$MINIKUBE" -eq 1 ]; then
  $REPO_ROOT/hack/make.py
  $REPO_ROOT/hack/dev/setup.sh
  my-operator run --v=4 \
    --secure-port=8443 \
    --kubeconfig="$HOME/.kube/config" \
    --authorization-kubeconfig="$HOME/.kube/config" \
    --authentication-kubeconfig="$HOME/.kube/config"
fi

if [ "$SELF_HOSTED" -eq 1 ]; then
  $REPO_ROOT/hack/docker/mg-operator/make.sh build
  $REPO_ROOT/hack/docker/mg-operator/make.sh push

  $REPO_ROOT/hack/deploy/setup.sh --docker-registry=${DOCKER_REGISTRY}
fi
