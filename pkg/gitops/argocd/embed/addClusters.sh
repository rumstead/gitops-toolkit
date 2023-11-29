#!/usr/bin/env bash
set -e
set -o pipefail

ARGO_PORT="${ARGOPORT:-"8080"}"
# support podman or any other non-docker gateway
CRI_GATEWAY="${CRI_GATEWAY:-"host.docker.internal"}"

# login
# https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
argocd login "$CRI_GATEWAY:$ARGO_PORT" --insecure $ARGOFLAGS --username "$ARGOUSER" --password "$ARGOPASSWD"

# don't quote $1 so it globs
argocd cluster add -y --upsert "$CONTEXT" --insecure $ARGOFLAGS --name "$CLUSTER" --kubeconfig "$KUBECONFIG" $1