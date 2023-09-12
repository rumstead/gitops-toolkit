# gitops-toolkit
Helpful manifests, scripts, and tools for gitops. Currently, the go binary supports creating N of clusters and allowing a GitOps engine to manage them.
The first use case was to set up an environment to test [Argo CD Application Sets](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/#introduction-to-applicationset-controller),
specifically with [cluster generators](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Generators-Cluster/).

## Troubleshooting
See the [troubleshooting guide](TROUBLESHOOTING.md).

## Shell Script
The first version of the toolkit was done via [shell scripts](hack/multiple-clusters/README.md).

## Getting Started
### Build/Install
```shell
make build
./bin/gitops-toolkit
```
```shell
go install github.com/rumstead/gitops-toolkit
```
### Configuring
The script reads a json/yaml configuration file to create clusters. Order of the clusters matters because of how K3d updates DNS. The bottom cluster can
address all cluster above it.
#### Schema
A json schema file can be found [here](pkg/config/v1alpha1/schema.json) with a [sample](pkg/config/testdata/clusters.json). Similar a yaml example file is [here](pkg/config/testdata/clusters.yaml).
#### Generating a configuration file
You can use the [proto structs](pkg/config/v1alpha1/cluster-config.pb.go) to write your configuration in code and dump them out as json.
## What is happening under the covers?

### Creates clusters
K3d is the only supported Kubernetes distribution.
```shell
k3d cluster list    
NAME    SERVERS   AGENTS   LOADBALANCER
admin   1/1       0/0      true
dev     1/1       0/0      true
qa      1/1       0/0      true
tst     1/1       0/0      true
```

### Deploys GitOps Engine
Argo CD is deployed to any configured GitOps clusters.
```shell
kgd -n argocd     
NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
argocd-redis                       1/1     1            1           129m
argocd-notifications-controller    1/1     1            1           129m
argocd-applicationset-controller   1/1     1            1           130m
argocd-repo-server                 1/1     1            1           129m
argocd-server                      1/1     1            1           129m
argocd-dex-server                  1/1     1            1           130m
```

### Link clusters to GitOps Engine
Argo CD is the only supported GitOps Engine.
```shell
kgsec -n  argocd --show-labels -l argocd.argoproj.io/secret-type=cluster 
NAME                                    TYPE     DATA   AGE    LABELS
cluster-k3d-admin-serverlb-2295321533   Opaque   3      118m   argocd.argoproj.io/secret-type=cluster
cluster-k3d-dev-serverlb-422902893      Opaque   3      119m   argocd.argoproj.io/secret-type=cluster,kubernetes.cnp.io/cluster.jurisdiction=k3d,
kubernetes.cnp.io/cluster.name=dev,kubernetes.cnp.io/cluster.region=muse2,kubernetes.cnp.io/cluster.segment=multitenant,kubernetes.cnp.io/environment=dev
cluster-k3d-tst-serverlb-756887653      Opaque   3      119m   argocd.argoproj.io/secret-type=cluster,kubernetes.cnp.io/cluster.jurisdiction=k3d,
kubernetes.cnp.io/cluster.name=tst,kubernetes.cnp.io/cluster.region=muse2,kubernetes.cnp.io/cluster.segment=multitenant,kubernetes.cnp.io/environment=tst
cluster-k3d-qa-serverlb-3703346418      Opaque   3      119m   argocd.argoproj.io/secret-type=cluster,kubernetes.cnp.io/cluster.jurisdiction=k3d,
kubernetes.cnp.io/cluster.name=qa,kubernetes.cnp.io/cluster.region=musw2,kubernetes.cnp.io/cluster.segment=multitenant,kubernetes.cnp.io/environment=qa
```