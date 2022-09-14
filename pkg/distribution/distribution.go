package distribution

import "github.com/rumstead/argo-cd-toolkit/pkg/distribution/k3d"

func NewCluster() Cluster {
	return &k3d.Cluster{}
}
