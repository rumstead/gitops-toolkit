package argocd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/gitops-toolkit/pkg/kubernetes"
)

func TestGenerateArgs(t *testing.T) {
	got := generateArgs(clusterArgLabels, map[string]string{"env": "dev"})
	if got != "--label env=dev " {
		t.Errorf("expected %q, got %q", "--label env=dev ", got)
	}

	got = generateArgs(clusterArgAnnotations, map[string]string{"a": "1", "b": "2"})
	for _, want := range []string{"--annotation a=1 ", "--annotation b=2 "} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q to contain %q", got, want)
		}
	}

	if got := generateArgs(clusterArgLabels, nil); got != "" {
		t.Errorf("expected empty string for nil metadata, got %q", got)
	}
}

func TestReplaceClusterUrl(t *testing.T) {
	kubeconfig := "apiVersion: v1\nclusters:\n- cluster:\n    server: https://0.0.0.0:40615\n  name: k3d-dev\n"
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(path, []byte(kubeconfig), 0600); err != nil {
		t.Fatal(err)
	}

	if err := replaceClusterUrl(path, "k3d-dev"); err != nil {
		t.Fatalf("replaceClusterUrl: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "server: https://k3d-dev-serverlb:6443") {
		t.Errorf("expected server url to be replaced, got:\n%s", got)
	}
	if strings.Contains(string(got), "0.0.0.0") {
		t.Errorf("expected 0.0.0.0 to be gone, got:\n%s", got)
	}
}

func TestGetBindAddress(t *testing.T) {
	agent := &Agent{}
	defaultCluster := &kubernetes.Cluster{RequestCluster: &v1alpha1.RequestCluster{GitOps: &v1alpha1.GitOps{}}}
	if got := agent.getBindAddress(defaultCluster); got != "0.0.0.0" {
		t.Errorf("expected default bind address 0.0.0.0, got %q", got)
	}

	custom := &kubernetes.Cluster{RequestCluster: &v1alpha1.RequestCluster{GitOps: &v1alpha1.GitOps{BindAddress: "localhost"}}}
	if got := agent.getBindAddress(custom); got != "localhost" {
		t.Errorf("expected bind address localhost, got %q", got)
	}
}
