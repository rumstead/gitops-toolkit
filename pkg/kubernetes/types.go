package kubernetes

import (
	"context"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
)

type Distro interface {
	CreateClusters(ctx context.Context, clusters *v1alpha1.RequestClusters) ([]*Cluster, error)
}

type Cluster struct {
	Name           string
	KubeConfigPath string
	*v1alpha1.RequestCluster
}
