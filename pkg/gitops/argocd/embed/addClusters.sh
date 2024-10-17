#!/usr/bin/env bash
set -ex

ARGO_PORT="${ARGOPORT:-"8080"}"
# support podman or any other non-docker gateway
CRI_GATEWAY="${CRI_GATEWAY:-"host.docker.internal"}"

# login
# https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
yes | argocd login "$CRI_GATEWAY:$ARGO_PORT" $ARGOFLAGS --username "$ARGOUSER" --password "$ARGOPASSWD"

# don't quote $1 so it globs
yes | argocd cluster add -y --upsert "$CONTEXT" $ARGOFLAGS --name "$CLUSTER" --kubeconfig "$KUBECONFIG" $1
