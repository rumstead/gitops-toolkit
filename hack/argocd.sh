#!/usr/bin/env bash
set -e
set -o pipefail

prompt() {
  echo "Press any key to continue... "
  read -r
}

precheck() {
    which docker > /dev/null || { echo "docker not installed or not on path"; exit; }
    which kubectl > /dev/null || { echo "kubectl not installed or not on path"; exit; }
    which argocd > /dev/null || { echo "argocd not installed or not on path"; exit; }
    which k3d > /dev/null || { echo "k3d not installed or not on path"; exit; }
    kc=$(kubectl config current-context)
    safe=$(echo "$kc" | grep -E -v "docker-desktop|minikube|rancher-desktop|k3d-*|kind-*" || true)
    if [ -n "$safe" ]
    then
      echo "You are using kube-context $safe which doesn't look like a local cluster"
      prompt
    fi
    echo "precheck passed"
}

# shout-out @bradfordwagner
deployArgoCD() {
  ns="$1"
  port="$2"
  manifestDir=$(find .. -name "argo-cd" -print -quit)
  kubectl create ns "$ns" || true
  kubectl apply -n "$ns" -k "$manifestDir"

  echo awaiting argocd server + redis to startup
  kubectl wait -n "$ns" deploy/argocd-server --for condition=available --timeout=5m
  kubectl wait -n "$ns" deploy/argocd-redis  --for condition=available --timeout=5m

  echo port forwarding argocd server
  kubectl port-forward -n "$ns" deploy/argocd-server "$port":8080 2>&1 > /dev/null &
  sleep 3

  # setup new password for argocd
  argo_host=localhost:${port}
  initial_password=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
  echo "https://${argo_host}" | pbcopy
  argocd login ${argo_host} \
    --username admin \
    --password "${initial_password}" \
    --insecure
  argocd account update-password \
    --account admin \
    --current-password "${initial_password}" \
    --new-password admin1234

 echo "Access the UI at: http://localhost:$port with user: admin and password: admin1234"
}

addClustersToArgoCD() {
  ARGOUSER="admin" ARGOPASSWD="admin1234" docker run --network localclusters --rm -e ARGOUSER -e ARGOPASSWD -v "$(pwd)/container:/hack" quay.io/argoproj/argocd:latest /hack/addClusters.sh
}