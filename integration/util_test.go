package integration_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
)

const defaultPort = 6379

func runServer(t *testing.T) func() {
	rootDir, err := runCommandAndCaptureOutput("git", "rev-parse", "--show-toplevel")
	if err != nil {
		t.Log(rootDir)
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

	// TODO: Uncomment once the server can handle many connections
	//if err := awaitServerStartup(); err != nil {
	//	t.Error(err.Error())
	//	stopServer(t, serverCmd)
	//}

	// TODO: remove once the server can handle many connections
	time.Sleep(2 * time.Second)

	return func() {
		stopServer(t, serverCmd)
	}
}

func awaitServerStartup() error {
	timeout := 5 * time.Second
	if !await.Until(serverIsUp, timeout, 200*time.Millisecond) {
		return fmt.Errorf("server did not start up in %s", timeout)
	}
	return nil
}

func serverIsUp() bool {
	_, err := dialServer()
	return err == nil
}

func dialServer() (net.Conn, error) {
	result, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", defaultPort))
	if err != nil {
		return nil, err
	}

	if err := result.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, err
	}

	return result, nil
}

func stopServer(t *testing.T, serverCmd *exec.Cmd) {
	if err := serverCmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("server did not stop gracefully: %v", err)
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
