package argocd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k3d-io/k3d/v5/pkg/logger"

	tkexec "github.com/rumstead/gitops-toolkit/pkg/exec"
	"github.com/rumstead/gitops-toolkit/pkg/kubernetes"

	_ "embed"

	"github.com/rumstead/gitops-toolkit/pkg/gitops"
	"github.com/rumstead/gitops-toolkit/pkg/logging"
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
	logging.Log().Infoln("Deploying Argo CD")
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
		// we don't want to error out if the namespace already exists
		if !strings.Contains(output, "already exists") {
			return fmt.Errorf("error creating namespace: %s: %v", output, err)
		} else {
			logging.Log().Infof("using the existing namespace: %s\n", ops.GetGitOps().GetNamespace())
		}
	}
	// 1a. wait for the cluster to be ready
	logging.Log().Debugln("waiting for cluster to be ready")
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", "kube-system", "deploy/coredns", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for cluster: %s: %v", output, err)
	}

	logging.Log().Debugln("deploying argo cd")
	// 2. apply the manifests
	cmd = exec.Command(a.cmd.Kubectl, "apply", "-n", ops.GetGitOps().GetNamespace(), "-k", ops.GetGitOps().GetManifestPath())
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error applying argo cd manifests at %s: %s: %v", ops.GetGitOps().GetManifestPath(), output, err)
	}
	logging.Log().Debugln("waiting for argo server and redis start up")
	// 3. wait for start up
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-server", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for argo server to be ready: %s: %v", output, err)
	}
	logging.Log().Debugln("argo server started")
	cmd = exec.Command(a.cmd.Kubectl, "wait", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-redis", "--for", "condition=available", "--timeout", "5m")
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error waiting for redis to be ready: %s: %v", output, err)
	}
	logging.Log().Debugln("redis started")

	logging.Log().Debugf("port forwarding on %s\n", ops.GetGitOps().GetPort())
	// 4. port forward
	// if port forward is not required return nil now
	if ops.GetGitOps().GetNoPortForward() {
		logging.Log().Infoln("Port forward is not required")
		return nil
	}
	bindAddress := a.getBindAddress(ops)
	port := fmt.Sprintf("%s:8080", ops.GetGitOps().GetPort())
	cmd = exec.Command(a.cmd.Kubectl, "port-forward", "-n", ops.GetGitOps().GetNamespace(), "deploy/argocd-server", port, "--address", bindAddress)
	// use start because we do not want to wait for the process to finish
	pid, err := tkexec.StartCommand(cmd)
	if err != nil {
		return fmt.Errorf("could not port foward argo server: %v", err)
	}
	logging.Log().Infof("port forward pid=%d,bind_addr=%s\n", pid, bindAddress)
	logging.Log().Infoln("argo cd deployed")
	return nil
}

func (a *Agent) getBindAddress(ops *kubernetes.Cluster) (bindAddress string) {
	// pull bind address from yaml config or default to 0.0.0.0 (maintaining backwards compatibility)
	bindAddress = "0.0.0.0"
	if ops.GetGitOps().GetBindAddress() != "" {
		bindAddress = ops.GetGitOps().GetBindAddress()
	}
	return
}

func (a *Agent) setAdminPassword(ops *kubernetes.Cluster) error {
	password, err := a.getInitialPassword(ops)
	if err != nil {
		return err
	}

	// yes command - this will auto confirm the unknown ca x509 error
	yesStdout, err := a.yesOut()
	if err != nil {
		return err
	}

	host := fmt.Sprintf("%s:%s", a.getBindAddress(ops), ops.GetGitOps().GetPort())
	// login
	args := []string{"login", host, "--username", ops.GetGitOps().GetCredentials().GetUsername(), "--password", password}
	logger.Log().Infof("logging into argo cd using the initial password args='%s'", strings.Join(args, " "))
	cmd := exec.Command(a.cmd.ArgoCD, args...)
	cmd.Stdin = yesStdout
	if output, err := tkexec.RunCommand(cmd); err != nil {
		logger.Log().Infoln("unable to log into argo cd using the initial password, trying config password: ", err.Error())
		logger.Log().Infoln("output", output)

		cmd = exec.Command(a.cmd.ArgoCD, "login", host, "--username", ops.GetGitOps().GetCredentials().GetUsername(), "--password", ops.GetGitOps().GetCredentials().GetPassword())
		cmd.Stdin = yesStdout
		if output, err := tkexec.RunCommand(cmd); err != nil {
			return fmt.Errorf("unable to log into argo cd %s: %v", output, err)
		}
	} else {
		args = []string{"account", "update-password", "--account", ops.GetGitOps().GetCredentials().GetUsername(), "--current-password",
			password, "--new-password", ops.GetGitOps().GetCredentials().GetPassword()}
		logger.Log().Infof("changing password args='%s'", strings.Join(args, " "))
		// change password
		cmd = exec.Command(a.cmd.ArgoCD, args...)
		cmd.Stdin = yesStdout
		if output, err := tkexec.RunCommand(cmd); err != nil {
			return fmt.Errorf("error changing argo cd password: %s: %v", output, err)
		}
	}
	logging.Log().Infof("access the UI at: http://%s user: %s password: %s\n", host, ops.GetGitOps().GetCredentials().GetUsername(), ops.GetGitOps().GetCredentials().GetPassword())
	return nil
}

func (a *Agent) yesOut() (io.ReadCloser, error) {
	yesCmd := exec.Command("yes")
	pipe, err := yesCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = yesCmd.Start()
	return pipe, err
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
	argoPort := fmt.Sprintf("ARGOPORT=%s", ops.GetGitOps().GetPort())
	contextName := fmt.Sprintf("CONTEXT=%s", workload.Name)
	clusterName := fmt.Sprintf("CLUSTER=%s", workload.RequestCluster.GetName())
	labels := generateArgs(clusterArgLabels, workload.GetLabels())
	annotations := generateArgs(clusterArgAnnotations, workload.GetAnnotations())
	cmd := exec.Command(a.cmd.CR, "run", "--network", ops.GetNetwork(), "--rm",
		"-e", argoUser,
		"-e", argoPasswd,
		"-e", argoPort,
		"-e", kubeConfig,
		"-e", contextName,
		"-e", clusterName,
		"-e", "CRI_GATEWAY",
		"-e", "ARGOFLAGS",
		"-v", workDirVolume,
		"quay.io/argoproj/argocd:latest", "/hack/addCluster.sh", labels+annotations)
	logging.Log().Debugf("%s\n", cmd.String())
	if output, err := tkexec.RunCommand(cmd); err != nil {
		return fmt.Errorf("error adding cluster to gitops agent: %s: %v", output, err)
	}
	logging.Log().Infof("added cluster %s to argo cd", workload.RequestCluster.GetName())
	return nil
}

func generateArgs(argType clusterArgs, metadata map[string]string) string {
	var builder strings.Builder
	for k, v := range metadata {
		builder.WriteString(string(argType))
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
