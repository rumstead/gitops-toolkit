#!/usr/bin/env bash
set -e
set -o pipefail

source clusters.sh
source argocd.sh

# create the clusters. the order matters here because of core dns entries
# we want the "central" cluster to be created last so it can target all the "workload" clusters
createCluster "dev"
createCluster "tst"
createCluster "qa"
createCluster "admin"
# make sure we have the proper binaries on our path and arent using a scary kubectx
precheck
namespace=argocd
port=8080
# order of the above clusters matter
# TODO: switch to the admin kubeconfig
deployArgoCD "$namespace" "$port"
addClustersToArgoCD