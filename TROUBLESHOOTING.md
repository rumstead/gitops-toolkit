# FAQ/Troubleshooting

## I want to pass in different K3s/K3d args
You can pass in any `k3s` argument or any `k3d` argument via the `additionalArgs` array. 

It is a great way to pass in a different k8s version. 

## level=fatal msg="dial tcp: lookup host.docker.internal..."
You can control the container gateway via the `CRI_GATEWAY` environment variable. 
Ie, for podman `CRI_GATEWAY=containers`
