package integration_test

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/await"
)

const defaultPort = 6379

func runServer(t *testing.T) {
	// TODO: strongly consider calling redis.RunServer() rather than starting
	//       the binary to save time and make things easier to debug.
	// TODO: if doing the above, keep one test around that does a smoke test
	//       against the binary.

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
		stopServer(t, serverCmd)
	})

	if err := awaitServerStartup(); err != nil {
		t.Error(err.Error())
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

	if err := result.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return nil, err
	}

	return result, nil
}

func mustDialServer(t *testing.T) net.Conn {
	conn, err := dialServer()
	if err != nil {
		t.Fatalf("no connection to server: %v", err)
	}
	t.Cleanup(func() {
		loggingClose(t, conn)
	})
	return conn
}

func stopServer(t *testing.T, serverCmd *exec.Cmd) {
	// TODO: Replace with serverCmd.Process.Signal(os.Interrupt) when the
	//       server can gracefully shut down:
	//   - https://victoriametrics.com/blog/go-graceful-shutdown/
	//   - https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
	if err := serverCmd.Process.Kill(); err != nil {
		t.Fatalf("server did not stop gracefully: %v", err)
	}
}

func doneChan(wg *sync.WaitGroup) chan struct{} {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	return done
}

func readResponse(conn net.Conn, responseLength int) (string, error) {
	gotBytes := make([]byte, responseLength)
	if _, err := io.ReadFull(conn, gotBytes); err != nil {
		return "", err
	}
	got := string(gotBytes)
	return got, nil
}

func loggingClose(t *testing.T, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		t.Log(err)
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
