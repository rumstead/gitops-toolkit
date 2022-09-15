package argocd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tkexec "github.com/rumstead/argo-cd-toolkit/pkg/exec"

	"github.com/rumstead/argo-cd-toolkit/pkg/config/v1alpha1"
	"github.com/rumstead/argo-cd-toolkit/pkg/gitops"
	"github.com/rumstead/argo-cd-toolkit/pkg/logging"
)

type Agent struct {
	kubectl string
	argocd  string
	cr      string
}

func NewGitOpsEngine(binaries map[string]string) gitops.Engine {
	return &Agent{
		kubectl: binaries["kubectl"],
		argocd:  binaries["argocd"],
		cr:      binaries["docker"],
	}
}

func (a *Agent) Deploy(ops *v1alpha1.GitOps, kubeconfig string) error {
	if _, err := os.Stat(kubeconfig); err != nil {
		return err
	}

	if err := a.deployArgoCD(ops); err != nil {
		return err
	}

	if err := a.setAdminPassword(ops); err != nil {
		return err
	}

	return nil
}

func (a *Agent) deployArgoCD(ops *v1alpha1.GitOps) error {
	logging.Log().Debugf("creating namespace: %s\n", ops.GetNamespace())
	cmd := exec.Command(a.kubectl, "create", "ns", ops.GetNamespace())
	// 1. create the ns
	if output, err := tkexec.RunCommand(cmd); err != nil {
		outputStr := string(output)
		// we don't want to error out if the namespace already exists
		if !strings.Contains(outputStr, "already exists") {
			return fmt.Errorf("error creating namespace: %s: %v", outputStr, err)
		} else {
			logging.Log().Infof("using the existing namespace: %s\n", ops.GetNamespace())
		}
	}
	// 1a. wait for the cluster to be ready
	logging.Log().Debugln("waiting for cluster to be ready")
	cmd = exec.Command(a.kubectl, "wait", "-n", "kube-system", "job/helm-install-traefik", "--for", "condition=complete", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for cluster: %s: %v", string(output), err)
	}

	logging.Log().Debugln("deploying argo cd")
	// 2. apply the manifests
	cmd = exec.Command(a.kubectl, "apply", "-n", ops.GetNamespace(), "-k", ops.GetManifestPath())
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error applying argo cd manifests at %s: %s: %v", ops.ManifestPath, string(output), err)
	}
	logging.Log().Debugln("waiting for argo server and redis start up")
	// 3. wait for start up
	cmd = exec.Command(a.kubectl, "wait", "-n", ops.GetNamespace(), "deploy/argocd-server", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for argo server to be ready: %s: %v", string(output), err)
	}
	logging.Log().Debugln("argo server started")
	cmd = exec.Command(a.kubectl, "wait", "-n", ops.GetNamespace(), "deploy/argocd-redis", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for redis to be ready: %s: %v", string(output), err)
	}
	logging.Log().Debugln("redis started")

	logging.Log().Debugf("port forwarding on %s\n", ops.GetPort())
	// 4. port forward
	// use start because we do not want to wait
	port := fmt.Sprintf("%s:8080", ops.GetPort())
	cmd = exec.Command(a.kubectl, "port-forward", "-n", ops.GetNamespace(), "deploy/argocd-server", port)
	pid, err := tkexec.StartCommand(cmd)
	if err != nil {
		return fmt.Errorf("could not port foward argo server: %v", err)
	}
	logging.Log().Infof("port forward pid: %d\n", pid)
	logging.Log().Infoln("argo cd deployed")
	return nil
}

func (a *Agent) setAdminPassword(ops *v1alpha1.GitOps) error {
	password, err := a.getInitialPassword(ops)
	if err != nil {
		return err
	}

	host := fmt.Sprintf("localhost:%s", ops.GetPort())
	// login
	cmd := exec.Command(a.argocd, "login", host, "--username", ops.GetCredentials().GetUsername(), "--password", password, "--insecure")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error logging into argo cd: %s: %v", string(output), err)
	}
	// change password
	cmd = exec.Command(a.argocd, "account", "update-password", "--account", ops.GetCredentials().GetUsername(), "--current-password",
		password, "--new-password", ops.GetCredentials().GetPassword())
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error changing argo cd password: %s: %v", string(output), err)
	}
	logging.Log().Debugf("access the UI at: http://%s user: %s password: %s\n", host, ops.GetCredentials().GetUsername(), ops.GetCredentials().GetPassword())
	return nil
}

func (a *Agent) getInitialPassword(ops *v1alpha1.GitOps) (string, error) {
	passwordCmd := exec.Command(a.kubectl, "get", "-n", ops.GetNamespace(), "secret", "argocd-initial-admin-secret", "-o", "jsonpath=\"{.data.password}\"")
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

func (a *Agent) AddClusters(cluster *v1alpha1.Clusters) error {
	return nil
}
