package v1alpha1_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
)

func unmarshalTestData(t *testing.T, path string, unmarshal func([]byte, interface{}) error) *v1alpha1.RequestClusters {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var clusters v1alpha1.RequestClusters
	if err := unmarshal(data, &clusters); err != nil {
		t.Fatalf("unmarshaling %s: %v", path, err)
	}
	return &clusters
}

func assertClusters(t *testing.T, clusters *v1alpha1.RequestClusters, wantBindAddress string) {
	t.Helper()
	names := []string{"dev", "tst", "qa", "admin"}
	if got := len(clusters.GetClusters()); got != len(names) {
		t.Fatalf("expected %d clusters, got %d", len(names), got)
	}
	for i, name := range names {
		cluster := clusters.GetClusters()[i]
		if cluster.GetName() != name {
			t.Errorf("cluster %d: expected name %q, got %q", i, name, cluster.GetName())
		}
		if cluster.GetNetwork() != "localclusters" {
			t.Errorf("cluster %q: expected network localclusters, got %q", name, cluster.GetNetwork())
		}
		if len(cluster.GetEnvs()) == 0 {
			t.Errorf("cluster %q: expected envs to be set", name)
		}
	}
	admin := clusters.GetClusters()[3]
	gitOps := admin.GetGitOps()
	if gitOps == nil {
		t.Fatal("expected admin cluster to have gitOps config")
	}
	if gitOps.GetNamespace() != "argocd" {
		t.Errorf("expected gitOps namespace argocd, got %q", gitOps.GetNamespace())
	}
	if gitOps.GetPort() != "8080" {
		t.Errorf("expected gitOps port 8080, got %q", gitOps.GetPort())
	}
	if gitOps.GetBindAddress() != wantBindAddress {
		t.Errorf("expected gitOps bindAddress %q, got %q", wantBindAddress, gitOps.GetBindAddress())
	}
	if gitOps.GetCredentials().GetUsername() != "admin" {
		t.Errorf("expected gitOps username admin, got %q", gitOps.GetCredentials().GetUsername())
	}
}

func TestUnmarshalYAML(t *testing.T) {
	assertClusters(t, unmarshalTestData(t, "../testdata/clusters.yaml", yaml.Unmarshal), "localhost")
}

func TestUnmarshalJSON(t *testing.T) {
	assertClusters(t, unmarshalTestData(t, "../testdata/clusters.json", json.Unmarshal), "")
}
