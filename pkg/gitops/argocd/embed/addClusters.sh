#!/usr/bin/env bash
set -e
set -o pipefail

# support podman or any other non-docker gateway
CRI_GATEWAY="${CRI_GATEWAY:-"docker"}"
# login
# https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
argocd login "host.$CRI_GATEWAY.internal:8080" --insecure --username "$ARGOUSER" --password "$ARGOPASSWD"

# don't quote $1 so it globs
argocd cluster add -y --upsert "$CONTEXT" --insecure --name "$CLUSTER" --kubeconfig "$KUBECONFIG" $1