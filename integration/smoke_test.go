//go:build integration

package integration_test

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/iox"
	"github.com/codecrafters-io/redis-starter-go/redis/redistest"
)

func TestSmoke(t *testing.T) {
	runServerBinary(t)
	conn := mustDialServer(t)

	if _, err := io.WriteString(conn, redistest.PingLowercase); err != nil {
		t.Fatalf(
			"did not write message %q successfully: %v",
			redistest.PingLowercase,
			err,
		)
	}

	got, err := iox.ReadExactly(conn, len(redistest.Pong))
	if err != nil {
		t.Errorf("did not read conn successfully: %v", err)
	}
	want := redistest.Pong
	if got != want {
		t.Errorf(
			`PING request: got response %q, want %q`, got, want,
		)
	}
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
		//       server can gracefully shut down:
		//   - https://victoriametrics.com/blog/go-graceful-shutdown/
		//   - https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
		//   - https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
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
