package gitops

import "github.com/rumstead/argo-cd-toolkit/pkg/config/v1alpha1"

type Engine interface {
	Deploy(ops *v1alpha1.GitOps) error
	AddClusters(cluster *v1alpha1.Clusters) error
}
