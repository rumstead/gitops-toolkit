package gitops

import (
	"context"

	"github.com/rumstead/argo-cd-toolkit/pkg/kubernetes"
)

type Engine interface {
	Deploy(ctx context.Context, ops *kubernetes.Cluster) error
	AddClusters(ctx context.Context, ops *kubernetes.Cluster, workload []*kubernetes.Cluster) error
}
