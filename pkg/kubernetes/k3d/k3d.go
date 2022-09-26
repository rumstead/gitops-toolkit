package k3d

import (
	"context"
	"errors"
	"fmt"

	k3dcluster "github.com/k3d-io/k3d/v5/cmd/cluster"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	log "github.com/sirupsen/logrus"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/gitops-toolkit/pkg/kubernetes"
	"github.com/rumstead/gitops-toolkit/pkg/random"
)

type K3d struct {
	workdir string
}

var errorCreate = errors.New("unable to create k3d cluster")

func NewK3dDistro(workdir string) kubernetes.Distro {
	return &K3d{workdir: workdir}
}

func (k *K3d) getKubeConfig(ctx context.Context, cluster *v1alpha1.RequestCluster) (string, error) {
	output := fmt.Sprintf("%s/%s", k.workdir, cluster.GetName())
	_, err := k3dclient.KubeconfigGetWrite(ctx, runtimes.Docker, &types.Cluster{Name: cluster.GetName()}, output, &k3dclient.WriteKubeConfigOptions{
		UpdateExisting:       true,
		UpdateCurrentContext: true,
		OverwriteExisting:    false,
	})
	if err != nil {
		return "", err
	}
	return output, nil
}

func (k *K3d) CreateClusters(ctx context.Context, clusters *v1alpha1.RequestClusters) ([]*kubernetes.Cluster, error) {
	var k8sClusters []*kubernetes.Cluster
	if clusters == nil {
		return nil, fmt.Errorf("invalid clusters provided: %w", errorCreate)
	}
	for _, cluster := range clusters.GetClusters() {
		log.Debugf("Creating cluster %s", cluster.GetName())
		k8sCluster, err := k.createCluster(ctx, cluster)
		k8sClusters = append(k8sClusters, k8sCluster)
		if err != nil {
			return nil, err
		}
	}
	return k8sClusters, nil
}

func parseClusterCreateArgs(cluster *v1alpha1.RequestCluster) []string {
	var args []string
	name := cluster.GetName()
	if name == "" {
		name = random.String(5)
	}
	args = append(args, name)

	for k, v := range cluster.GetEnvs() {
		args = append(args, "-e")
		arg := fmt.Sprintf("%s=%s", k, v)
		args = append(args, arg)
	}

	if cluster.GetNetwork() != "" {
		args = append(args, "--network")
		args = append(args, cluster.GetNetwork())
	}

	for k, v := range cluster.GetVolumes() {
		args = append(args, "--volume")
		arg := fmt.Sprintf("%s:%s", k, v)
		args = append(args, arg)
	}

	for k, v := range cluster.GetAdditionalArgs() {
		args = append(args, "--k3s-arg")
		arg := fmt.Sprintf("%s=%s", k, v)
		args = append(args, arg)
	}
	return args
}

func (k *K3d) createCluster(ctx context.Context, cluster *v1alpha1.RequestCluster) (*kubernetes.Cluster, error) {
	if !clusterExists(ctx, cluster) {
		cmd := k3dcluster.NewCmdClusterCreate()
		args := parseClusterCreateArgs(cluster)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			return nil, err
		}
	} else {
		log.Warnf("cluster %s already exists", cluster.GetName())
	}
	config, err := k.getKubeConfig(ctx, cluster)
	if err != nil {
		return nil, err
	}
	clusterName := fmt.Sprintf("k3d-%s", cluster.GetName())
	return &kubernetes.Cluster{Name: clusterName, RequestCluster: cluster, KubeConfigPath: config}, nil
}

func clusterExists(ctx context.Context, cluster *v1alpha1.RequestCluster) bool {
	// check if a cluster with that name exists already
	if _, err := k3dclient.ClusterGet(ctx, runtimes.Docker, &types.Cluster{Name: cluster.GetName()}); err == nil {
		return true
	}
	return false
}
