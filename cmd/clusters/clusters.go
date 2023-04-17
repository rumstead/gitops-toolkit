/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusters

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ghodss/yaml"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/gitops-toolkit/pkg/gitops/argocd"
	"github.com/rumstead/gitops-toolkit/pkg/kubernetes"
	"github.com/rumstead/gitops-toolkit/pkg/kubernetes/k3d"
	"github.com/rumstead/gitops-toolkit/pkg/logging"
	"github.com/spf13/cobra"
)

var cfgFile string

var binaries = map[string]string{"k3d": "", "docker": "", "kubectl": "", "argocd": ""}

func NewClustersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Create a set of k3d clusters managed by Argo CD",
		Long:  ``,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			// validate args
			if _, err := os.Stat(cfgFile); err != nil {
				if os.IsNotExist(err) {
					logging.Log().Errorf("config file %s doesn't exist: %v", cfgFile, err)
					return err
				}
				return err
			}
			if err := checkPath(binaries); err != nil {
				logging.Log().Fatalf("PATH is missing binaries. %v", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			// TODO: Make timeout configurable
			timeoutCtx, timeoutFunc := context.WithTimeout(ctx, 20*time.Minute)
			defer timeoutFunc()
			data, err := os.ReadFile(cfgFile)
			if err != nil {
				return err
			}
			var requestedClusters v1alpha1.RequestClusters
			if err = yaml.Unmarshal(data, &requestedClusters); err != nil {
				logging.Log().Fatalf("unable to parse %s cluster config: %v", cfgFile, err)
			}

			outputDir, err := getOutputDir()
			if err != nil {
				return err
			}
			workdir := fmt.Sprintf("%s/gitops-toolkit/", outputDir)
			defer os.RemoveAll(workdir)
			// create the clusters
			clusterDistro := k3d.NewK3dDistro(workdir)
			k8sClusters, err := clusterDistro.CreateClusters(timeoutCtx, &requestedClusters)
			if err != nil {
				logging.Log().Fatalf("error creating clusters: %v", err)
			}

			// get any clusters to deploy gitops engine to
			var gitopsClusters []*kubernetes.Cluster
			for _, cluster := range k8sClusters {
				if cluster.GetGitOps() != nil {
					gitopsClusters = append(gitopsClusters, cluster)
				}
			}

			// deploy the gitops engine to any enabled clusters
			gitOpsEngine := argocd.NewGitOpsEngine(binaries)
			if err != nil {
				return err
			}

			for _, ops := range gitopsClusters {
				if err = gitOpsEngine.Deploy(ctx, ops); err != nil {
					logging.Log().Fatalf("error deploying gitops: %v", err)
				}

				if err = gitOpsEngine.AddClusters(ctx, ops, k8sClusters); err != nil {
					logging.Log().Fatalf("error adding cluster to gitops engine: %v", err)
				}
			}
			// can help if running in an IDE
			return nil
		},
	}
	defaultClusterConfigPath := getDefaultClusterConfig()
	cmd.PersistentFlags().StringVar(&cfgFile, "config", defaultClusterConfigPath, "path to a config file containing clusters")
	return cmd
}

func getDefaultClusterConfig() (filePath string) {
	homeDir, _ := os.UserHomeDir()
	filePath = fmt.Sprintf("%s/.gitops-toolkit-clusters.yaml", homeDir)
	return
}

func getOutputDir() (string, error) {
	dir := os.Getenv("OUTPUT_DIR")
	if dir == "" {
		return os.TempDir(), nil
	}
	return dir, nil
}

func checkPath(binaries map[string]string) error {
	for binary := range binaries {
		path, err := exec.LookPath(binary)
		if err != nil {
			return err
		}
		binaries[binary] = path
	}
	return nil
}
