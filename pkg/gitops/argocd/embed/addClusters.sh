#!/usr/bin/env bash
set -e
set -o pipefail

ARGO_PORT="${ARGOPORT:-"8080"}"
# support podman or any other non-docker gateway
CRI_GATEWAY="${CRI_GATEWAY:-"docker"}"

if [ -z "$DOCKERGATEWAY" ]
then
      DOCKER_GATEWAY="host.$CRI_GATEWAY.internal"
else
      DOCKER_GATEWAY="$DOCKERGATEWAY"
fi
# login
# https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
argocd login "$DOCKER_GATEWAY:$ARGO_PORT" --insecure $ARGOFLAGS --username "$ARGOUSER" --password "$ARGOPASSWD"

# don't quote $1 so it globs
argocd cluster add -y --upsert "$CONTEXT" --insecure $ARGOFLAGS --name "$CLUSTER" --kubeconfig "$KUBECONFIG" $1