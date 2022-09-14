package k3d

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"

	"github.com/rumstead/argo-cd-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/argo-cd-toolkit/pkg/random"
)

type Cluster struct {
}

var errorCreate = errors.New("unable to create k3d cluster")

func (k *Cluster) Create(clusters *v1alpha1.Clusters) error {
	if clusters == nil {
		return fmt.Errorf("invalid clusters provided: %w", errorCreate)
	}
	for _, cluster := range clusters.Cluster {
		if err := createCluster(cluster); err != nil {
			return err
		}
	}
	return nil
}

func parseClusterCreateArgs(cluster *v1alpha1.Cluster) []string {
	var args []string
	name := cluster.GetName()
	if name == "" {
		name = random.String(5)
	}
	args = append(args, name)

	for k, v := range cluster.GetEnvs() {
		arg := fmt.Sprintf("-e %s=%s", k, v)
		args = append(args, arg)
	}

	if cluster.GetNetwork() != "" {
		args = append(args, cluster.GetNetwork())
	}

	for k, v := range cluster.GetVolumes() {
		args = append(args, "--volume")
		arg := fmt.Sprintf("%s=%s", k, v)
		args = append(args, arg)
	}

	for k, v := range cluster.GetAdditionalArgs() {
		args = append(args, "--k3s-arg")
		arg := fmt.Sprintf("%s=%s", k, v)
		args = append(args, arg)
	}
	return args
}

func createCluster(cluster *v1alpha1.Cluster) error {
	if !clusterExists(cluster) {
		cmd := k3dcluster.NewCmdClusterCreate()
		args := parseClusterCreateArgs(cluster)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			return err
		}
	} else {
		log.Warnf("cluster %s already exists", cluster.GetName())
	}
	return nil
}

func clusterExists(cluster *v1alpha1.Cluster) bool {
	// check if a cluster with that name exists already
	if _, err := k3dclient.ClusterGet(context.Background(), runtimes.Docker, &types.Cluster{Name: cluster.GetName()}); err == nil {
		return true
	}
	return false
}
