# FAQ/Troubleshooting

## I want to pass in different K3s/K3d args
You can pass in any `k3s` argument or any `k3d` argument via the `additionalArgs` array.

It is a great way to pass in a different k8s version.

## level=fatal msg="dial tcp: lookup host.docker.internal..."
You can control the container gateway hostname via the `CRI_GATEWAY` environment variable. By default the container gateway hostname is `host.docker.internal`. Ie:
- for podman `CRI_GATEWAY=host.containers.internal`
- other hosts `CRI_GATEWAY=my-gateway`
- or ip `CRI_GATEWAY=172.18.0.1`

## Argo is behind a reverse proxy (ingress like treafik)
You can add required flags, such as `--grpc-web`, to the argocd commands by adding `ARGOFLAGS` an an environment variable.
Ie, `ARGOFLAGS=--grpc-web`