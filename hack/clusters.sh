#!/usr/bin/env bash
set -e
set -o pipefail

# Create cluster by name on the same docker network as all other clusters
# the network name is "localclusters"
# https://k3d.io/v5.4.6/faq/faq/#pods-fail-to-start-x509-certificate-signed-by-unknown-authority
createCluster() {
  if ! k3d cluster get "$1" > /dev/null; then
    k3d cluster create "$1"  \
      -e "http_proxy=$http_proxy@all" -e "HTTP_PROXY=$HTTP_PROXY@all" -e "no_proxy=$no_proxy@all" \
      -e "NO_PROXY=$NO_PROXY@all" -e "https_proxy=$http_proxy@all" -e "HTTPS_PROXY=$HTTP_PROXY@all" \
      --network localclusters \
      --volume "$HOME"/tmp/certs/internal-ca-bundle.crt:/etc/ssl/certs/corp.crt \
      --k3s-arg "--tls-san=k3d-$1-serverlb"@server:*
  else
    echo "cluster $1 already exists"
  fi
  kubeconfigDir=$(find . -name "container" -print -quit)
  kubeconfigFile="$kubeconfigDir/kubeconfig/$1.yaml"
  echo "dropping kubeconfig to $kubeconfigFile"
  mkdir -p "$kubeconfigDir/kubeconfig/"
  k3d kubeconfig get "$1" > "$kubeconfigFile"
  # sed... isn't very portable, this requires GNU sed
  sed -Ei "s/0.0.0.0:[0-9]+/k3d-$1-serverlb:6443/" "$kubeconfigFile"
}
