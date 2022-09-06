#!/usr/bin/env bash
set -e
set -o pipefail

# login
# https://docs.docker.com/desktop/networking/#i-want-to-connect-from-a-container-to-a-service-on-the-host
argocd login host.docker.internal:8080 --insecure --username "$ARGOUSER" --password "$ARGOPASSWD"

kubeconfigDir=$(find /hack -name "kubeconfig" -print -quit)
for kc in "$kubeconfigDir"/*; do
  cluster=$(basename "$kc" .yaml)
  argocd cluster add -y --upsert --kubeconfig "$kc" "k3d-$cluster" --insecure --name "$cluster"
done