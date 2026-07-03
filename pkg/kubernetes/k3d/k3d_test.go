package k3d

import (
	"slices"
	"testing"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
)

func TestParseClusterCreateArgs(t *testing.T) {
	cluster := &v1alpha1.RequestCluster{
		Name:           "dev",
		Network:        "localclusters",
		Envs:           map[string]string{"HTTP_PROXY": "@all"},
		Volumes:        map[string]string{"/tmp/ca.crt": "/etc/ssl/certs/corp.crt"},
		AdditionalArgs: []string{"--image=rancher/k3s:v1.26.4-k3s1"},
	}
	args := parseClusterCreateArgs(cluster)

	if args[0] != "dev" {
		t.Errorf("expected first arg to be cluster name, got %q", args[0])
	}
	for _, pair := range [][]string{
		{"-e", "HTTP_PROXY=@all"},
		{"--network", "localclusters"},
		{"--volume", "/tmp/ca.crt:/etc/ssl/certs/corp.crt"},
	} {
		i := slices.Index(args, pair[0])
		if i == -1 || i+1 >= len(args) || args[i+1] != pair[1] {
			t.Errorf("expected args to contain %q %q, got %v", pair[0], pair[1], args)
		}
	}
	if !slices.Contains(args, "--image=rancher/k3s:v1.26.4-k3s1") {
		t.Errorf("expected additional args to be passed through, got %v", args)
	}
}

func TestParseClusterCreateArgsGeneratesRandomName(t *testing.T) {
	args := parseClusterCreateArgs(&v1alpha1.RequestCluster{})
	if len(args) != 1 || len(args[0]) != 5 {
		t.Errorf("expected a single generated 5 character name, got %v", args)
	}
}
