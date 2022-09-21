package argocd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k3d-io/k3d/v5/pkg/logger"

	tkexec "github.com/rumstead/argo-cd-toolkit/pkg/exec"
	"github.com/rumstead/argo-cd-toolkit/pkg/kubernetes"

	_ "embed"

	"github.com/rumstead/argo-cd-toolkit/pkg/gitops"
	"github.com/rumstead/argo-cd-toolkit/pkg/logging"
)

//go:embed embed/addClusters.sh
var shellScript []byte

type Agent struct {
	cmd *tkexec.Command
}

func NewGitOpsEngine(binaries map[string]string) gitops.Engine {
	return &Agent{cmd: tkexec.NewCommand(binaries)}
}

func (a *Agent) Deploy(ctx context.Context, ops *kubernetes.Cluster) error {
	if _, err := os.Stat(ops.KubeConfigPath); err != nil {
		return err
	}

	oldKubeconfig := os.Getenv("KUBECONFIG")
	err := os.Setenv("KUBECONFIG", ops.KubeConfigPath)
	if err != nil {
		return err
	}
	defer os.Setenv("KUBECONFIG", oldKubeconfig)

	if err = a.deployArgoCD(ctx, ops); err != nil {
		return err
	}

	if err = a.setAdminPassword(ops); err != nil {
		return err
	}

	return nil
}

func (a *Agent) deployArgoCD(_ context.Context, ops *kubernetes.Cluster) error {
	logging.Log().Debugf("creating namespace: %s\n", ops.GetGitOps().GetNamespace())
	cmd := exec.Command(a.cmd.Kubectl, "create", "ns", ops.GetGitOps().GetNamespace())
	// 1. create the ns
	if output, err := tkexec.RunCommand(cmd); err != nil {
		outputStr := string(output)
		// we don't want to error out if the namespace already exists
		if !strings.Contains(outputStr, "already exists") {
			return fmt.Errorf("error creating namespace: %s: %v", outputStr, err)
		} else {
			logging.Log().Infof("using the existing namespace: %s\n", ops.GetGitOps().GetNamespace())
		}
	}
	// 1a. wait for the cluster to be ready
	logging.Log().Debugln("waiting for cluster to be ready")
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", "kube-system", "job/helm-install-traefik", "--for", "condition=complete", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for cluster: %s: %v", string(output), err)
	}

	logging.Log().Debugln("deploying argo cd")
	// 2. apply the manifests
	cmd = exec.Command(a.cmd.Kubectl, "apply", "-n", ops.GetGitOps().GetNamespace(), "-k", ops.GetGitOps().GetManifestPath())
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error applying argo cd manifests at %s: %s: %v", ops.GetGitOps().GetManifestPath(), string(output), err)
	}
	logging.Log().Debugln("waiting for argo server and redis start up")
	// 3. wait for start up
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-server", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for argo server to be ready: %s: %v", string(output), err)
	}
	logging.Log().Debugln("argo server started")
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-redis", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for redis to be ready: %s: %v", string(output), err)
	}
	logging.Log().Debugln("redis started")

	logging.Log().Debugf("port forwarding on %s\n", ops.GetGitOps().GetPort())
	// 4. port forward
	// use start because we do not want to wait
	port := fmt.Sprintf("%s:8080", ops.GetGitOps().GetPort())
	cmd = exec.Command(a.cmd.Kubectl, "port-forward", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-server", port)
	pid, err := tkexec.StartCommand(cmd)
	if err != nil {
		return fmt.Errorf("could not port foward argo server: %v", err)
	}
	logging.Log().Infof("port forward pid: %d\n", pid)
	logging.Log().Infoln("argo cd deployed")
	return nil
}

func (a *Agent) setAdminPassword(ops *kubernetes.Cluster) error {
	password, err := a.getInitialPassword(ops)
	if err != nil {
		return err
	}

	host := fmt.Sprintf("localhost:%s", ops.GetGitOps().GetPort())
	// login
	cmd := exec.Command(a.cmd.ArgoCD, "login", host, "--username", ops.GetGitOps().GetCredentials().GetUsername(), "--password", password, "--insecure")
	if _, err := tkexec.RunCommand(cmd); err != nil {
		logger.Log().Infoln("unable to log into argo cd using the initial password, trying config password")
		cmd = exec.Command(a.cmd.ArgoCD, "login", host, "--username", ops.GetGitOps().GetCredentials().GetUsername(), "--password", ops.GetGitOps().GetCredentials().GetPassword(), "--insecure")
		if output, err := tkexec.RunCommand(cmd); err != nil {
			return fmt.Errorf("unable to log into argo cd %s: %v", string(output), err)
		}
	} else {
		// change password
		cmd = exec.Command(a.cmd.ArgoCD, "account", "update-password", "--account", ops.GetGitOps().GetCredentials().GetUsername(), "--current-password",
			password, "--new-password", ops.GetGitOps().GetCredentials().GetPassword())
		if output, err := tkexec.RunCommand(cmd); err != nil {
			return fmt.Errorf("error changing argo cd password: %s: %v", string(output), err)
		}
	}
	logging.Log().Debugf("access the UI at: http://%s user: %s password: %s\n", host, ops.GetGitOps().GetCredentials().GetUsername(), ops.GetGitOps().GetCredentials().GetPassword())
	return nil
}

func (a *Agent) getInitialPassword(ops *kubernetes.Cluster) (string, error) {
	passwordCmd := exec.Command(a.cmd.Kubectl, "get", "-n", ops.GetGitOps().GetNamespace(), "secret", "argocd-initial-admin-secret", "-o", "jsonpath=\"{.data.password}\"")
	outputBytes, err := tkexec.RunCommandCaptureStdOut(passwordCmd)
	if err != nil {
		return "", fmt.Errorf("error getting argocd password %v", err)
	}
	password := string(outputBytes)
	// we need to trim the quotes from the base64 string
	bytePassword := []byte(strings.Trim(password, "\""))
	decodeBuf := make([]byte, base64.StdEncoding.DecodedLen(len(bytePassword)))
	if n, err := base64.StdEncoding.Decode(decodeBuf, bytePassword); err != nil {
		return "", fmt.Errorf("error decoding b64 argocd password: %d, %v", n, err)
	}
	// remove null
	decodeBuf = bytes.Trim(decodeBuf, "\x00")
	return string(decodeBuf), nil
}

func (a *Agent) AddClusters(ctx context.Context, ops *kubernetes.Cluster, workload []*kubernetes.Cluster) error {
	for _, cluster := range workload {
		err := a.AddCluster(ctx, ops, cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) AddCluster(_ context.Context, ops, workload *kubernetes.Cluster) error {
	if err := replaceClusterUrl(workload.KubeConfigPath, workload.Name); err != nil {
		return err
	}
	workdir := filepath.Dir(workload.KubeConfigPath)
	workDirVolume := fmt.Sprintf("%s/:%s/", workdir, "/hack")
	kubeConfig := fmt.Sprintf("KUBECONFIG=%s/%s", "/hack", workload.RequestCluster.GetName())
	addClusterPath := fmt.Sprintf("%s/%s", workdir, "addCluster.sh")

	if err := os.WriteFile(addClusterPath, shellScript, 0777); err != nil {
		return err
	}
	argoUser := fmt.Sprintf("ARGOUSER=%s", ops.GetGitOps().GetCredentials().GetUsername())
	argoPasswd := fmt.Sprintf("ARGOPASSWD=%s", ops.GetGitOps().GetCredentials().GetPassword())
	cluster := fmt.Sprintf("CLUSTER=%s", workload.Name)
	labels := generateLabels(workload.GetLabels())
	cmd := exec.Command(a.cmd.CR, "run", "--network", ops.GetNetwork(), "--rm", "-e", argoUser, "-e", argoPasswd, "-e", kubeConfig, "-e", cluster,
		"-v", workDirVolume, "quay.io/argoproj/argocd:latest", "/hack/addCluster.sh", labels)
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error adding cluster to gitops agent: %s: %v", string(output), err)
	}
	logging.Log().Infof("added cluster %s to argo cd\n", workload.Name)
	return nil
}

func generateLabels(labels map[string]string) string {
	var builder strings.Builder
	for k, v := range labels {
		builder.WriteString("--label")
		builder.WriteString(" ")
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(v)
		builder.WriteString(" ")
	}
	return builder.String()
}

func replaceClusterUrl(path, clusterName string) error {
	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// TODO: port configurable
	server := fmt.Sprintf("%s-serverlb:%s", clusterName, "6443")
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "0.0.0.0") {
			replacement := serverReplace.ReplaceAllString(line, server)
			lines[i] = replacement
		}
	}
	output := strings.Join(lines, "\n")
	err = os.WriteFile(path, []byte(output), 0777)
	if err != nil {
		return err
	}
	return nil
}
