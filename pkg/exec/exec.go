package exec

import (
	"io"
	"os/exec"

	"github.com/rumstead/gitops-toolkit/pkg/logging"
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

func readStdOut(out chan []byte, reader io.ReadCloser) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		logging.Log().Errorf("unable to read stdout of process %v", err)
		return
	}
	out <- buf
}

func RunCommandCaptureStdOut(cmd *exec.Cmd) ([]byte, error) {
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	out := make(chan []byte)
	go readStdOut(out, reader)

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return <-out, nil
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
