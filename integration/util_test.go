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
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	timeoutDuration := 5 * time.Second
	timeout := time.NewTimer(timeoutDuration)
	defer timeout.Stop()
	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("server did not start up in %s", timeoutDuration)
		case <-ticker.C:
			if serverIsUp() {
				// Success
				return nil
			}
		}
	}
}

func serverIsUp() bool {
	_, err := dialServer()
	return err == nil
}

func dialServer() (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("localhost:%d", defaultPort))
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
