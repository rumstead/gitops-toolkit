package distribution

import "github.com/rumstead/argo-cd-toolkit/pkg/config/v1alpha1"

type Cluster interface {
	Create(clusters *v1alpha1.Clusters) error
}
