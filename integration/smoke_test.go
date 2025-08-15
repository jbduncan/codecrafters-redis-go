package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestSmoke(t *testing.T) {
	request := redistest.RequestPingLowercase
	response := redistest.ResponsePong

	runServerBinary(t)
	conn := mustDialServer(t)

	redistest.TestRequestAndResponse(t, conn, request, response)
}

func runServerBinary(t *testing.T) {
	rootDir, err := runCommandAndCaptureOutput("git", "rev-parse", "--show-toplevel")
	if err != nil {
		t.Logf("git rev-parse --show-toplevel: %s", rootDir)
		t.Fatalf("server not built: %v", err)
	}

	serverBinaryPath := filepath.Join(t.TempDir(), "server")
	if err := cmd(
		"go",
		"build",
		"-o",
		serverBinaryPath,
		rootDir,
	).Run(); err != nil {
		t.Fatalf("server not built: %v", err)
	}

	serverCmd := cmd(serverBinaryPath)
	if err := serverCmd.Start(); err != nil {
		t.Fatalf("server did not start: %v", err)
	}

	t.Cleanup(func() {
		// TODO: Replace with serverCmd.Process.Signal(os.Interrupt) when the
		//       server can gracefully shut down.
		if err := serverCmd.Process.Kill(); err != nil {
			t.Fatalf("server did not stop gracefully: %v", err)
		}
	})

	if err := awaitServerStartUp(); err != nil {
		t.Error(err.Error())
	}
}

func runCommandAndCaptureOutput(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	output, err := c.CombinedOutput()
	return strings.TrimRight(string(output), "\n"), err
}

func cmd(name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c
}
