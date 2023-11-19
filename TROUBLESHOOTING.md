# FAQ/Troubleshooting

## I want to pass in different K3s/K3d args
You can pass in any `k3s` argument or any `k3d` argument via the `additionalArgs` array.

It is a great way to pass in a different k8s version.

## level=fatal msg="dial tcp: lookup host.docker.internal..."
You can control the container gateway via the `CRI_GATEWAY` environment variable.
Ie, for podman `CRI_GATEWAY=containers`

## Argo is behind a reverse proxy (ingress like treafik)
You can add required flags, such as `--grpc-web`, to the argocd commands by adding `ARGOFLAGS` an an environment variable.
Ie, `ARGOFLAGS=--grpc-web`

## Are you using wsl and host.docker.internal does not work?
You can pass a host name or IP address via the `ARGOHOST` environment variable.
Ie, `ARGOHOST=my-host` or `ARGOHOST=172.18.0.1`