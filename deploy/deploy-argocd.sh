#!/bin/sh
set -e
set -o pipefail

prompt() {
  echo "Press any key to continue... "
  read -r
}

precheck() {
    which docker > /dev/null || { echo "docker not install or not on path"; exit; }
    which kubectl > /dev/null || { echo "kubectl not install or not on path"; exit; }
    which argocd > /dev/null || { echo "argocd not install or not on path"; exit; }
    kc=$(kubectl config current-context)
    safe=$(echo "$kc" | grep -E -v "docker-desktop|minikube|rancher-desktop|k3s-*|kind-*" || true)
    if [ -n "$safe" ]
    then
      echo "You are using kube-context $safe which doesn't look like a local cluster"
      prompt
    fi
    echo "precheck passed"
}

# shout-out @bradfordwagner
deploy() {
  ns=argocd
  kubectl create ns ${ns}
  kubectl apply -n ${ns} -k "./manifests/"

  echo awaiting argocd server + redis to startup
  kubectl wait -n ${ns} deploy/argocd-server --for condition=available --timeout=5m
  kubectl wait -n ${ns} deploy/argocd-redis  --for condition=available --timeout=5m

  echo port forwarding argocd server
  port_forward=8080
  kubectl port-forward -n ${ns} deploy/argocd-server ${port_forward}:8080 2>&1 > /dev/null &
  sleep 3

  # setup new password for argocd
  argo_host=localhost:${port_forward}
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

 echo "Access the UI at: http://localhost:8080 with user: admin and password: admin1234"
}

precheck
deploy