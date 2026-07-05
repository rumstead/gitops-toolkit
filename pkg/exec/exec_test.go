package exec

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(map[string]string{
		"kubectl": "/usr/bin/kubectl",
		"argocd":  "/usr/bin/argocd",
		"docker":  "/usr/bin/docker",
	})
	if cmd.Kubectl != "/usr/bin/kubectl" || cmd.ArgoCD != "/usr/bin/argocd" || cmd.CR != "/usr/bin/docker" {
		t.Errorf("unexpected command mapping: %+v", cmd)
	}
}

func TestRunCommandCaptureStdOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test relies on unix shell utilities")
	}
	out, err := RunCommandCaptureStdOut(exec.Command("echo", "hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("expected %q, got %q", "hello\n", string(out))
	}
}

func TestRunCommandCaptureStdOutError(t *testing.T) {
	if _, err := RunCommandCaptureStdOut(exec.Command("/nonexistent-binary")); err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestRunCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test relies on unix shell utilities")
	}
	output, err := RunCommand(exec.Command("true"))
	if err != nil || output != "" {
		t.Errorf("expected success with no output, got output=%q err=%v", output, err)
	}

	output, err = RunCommand(exec.Command("sh", "-c", "echo boom; exit 1"))
	if err == nil {
		t.Fatal("expected error from failing command")
	}
	if output != "boom\n" {
		t.Errorf("expected combined output %q, got %q", "boom\n", output)
	}
}
