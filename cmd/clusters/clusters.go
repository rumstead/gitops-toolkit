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
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/rumstead/argo-cd-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/argo-cd-toolkit/pkg/distribution"
)

var cfgFile string

var binaries = []string{"k3d", "docker", "kubectl", "argocd"}

func NewClustersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			// validate args
			if _, err := os.Stat(cfgFile); err != nil {
				if os.IsNotExist(err) {
					log.Fatalf("config file %s doesn't exist: %v", cfgFile, err)
				}
				return err
			}
			if err := checkPath(binaries); err != nil {
				log.Fatalf("PATH is missing binaries. %v", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(cfgFile)
			if err != nil {
				return err
			}
			var clusters v1alpha1.Clusters
			if err = protojson.Unmarshal(data, &clusters); err != nil {
				log.Fatalf("unable to parse %s cluster config: %v", cfgFile, err)
			}
			if err = distribution.NewCluster().Create(&clusters); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to a config file containing clusters")
	return cmd
}

func checkPath(binaries []string) error {
	for _, binary := range binaries {
		if _, err := exec.LookPath(binary); err != nil {
			return err
		}
	}
	return nil
}
