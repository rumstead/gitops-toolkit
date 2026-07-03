package exec

import (
	"bytes"
	"os/exec"
)

type Command struct {
	Kubectl string
	ArgoCD  string
	CR      string
}

func NewCommand(binaries map[string]string) *Command {
	return &Command{
		Kubectl: binaries["kubectl"],
		ArgoCD:  binaries["argocd"],
		CR:      binaries["docker"],
	}
}

func RunCommandCaptureStdOut(cmd *exec.Cmd) ([]byte, error) {
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func RunCommand(cmd *exec.Cmd) (string, error) {
	if output, err := cmd.CombinedOutput(); err != nil {
		return string(output), err
	}
	return "", nil
}

func StartCommand(cmd *exec.Cmd) (int, error) {
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return -1, err
	}
	return pid, nil
}
