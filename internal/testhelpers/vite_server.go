package testhelpers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ozanturksever/logutil"
)

// ViteServer represents a Vite development server for testing
type ViteServer struct {
	exampleName string
	host        string
	port        int
	cmd         *exec.Cmd
	cancel      context.CancelFunc
	url         string
}

// NewViteServer creates a new Vite server instance for the given example
func NewViteServer(exampleName, address string) *ViteServer {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		panic(fmt.Sprintf("Invalid address format: %v", err))
	}

	port := 0
	if portStr != "0" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			panic(fmt.Sprintf("Invalid port: %v", err))
		}
	}

	return &ViteServer{
		exampleName: exampleName,
		host:        host,
		port:        port,
	}
}

// Start starts the Vite development server
func (s *ViteServer) Start() error {
	// Find an available port if port is 0
	if s.port == 0 {
		listener, err := net.Listen("tcp", s.host+":0")
		if err != nil {
			return fmt.Errorf("failed to find available port: %v", err)
		}
		s.port = listener.Addr().(*net.TCPAddr).Port
		listener.Close()
	}

	s.url = fmt.Sprintf("http://%s:%d", s.host, s.port)

	// Create context for the command
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Get the project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to find project root: %v", err)
	}

	// Build the npm command to start Vite for the specific example
	cmdArgs := []string{"run", fmt.Sprintf("dev:%s", s.exampleName), "--", "--port", strconv.Itoa(s.port), "--host", s.host}
	s.cmd = exec.CommandContext(ctx, "npm", cmdArgs...)
	s.cmd.Dir = projectRoot

	// Set environment to avoid npm warnings
	s.cmd.Env = append(os.Environ(), "NODE_ENV=development")

	// Determine debug mode from env var
	debugEnv := os.Getenv("UIWGO_VITE_DEBUG")
	debug := strings.EqualFold(debugEnv, "1") || strings.EqualFold(debugEnv, "true") || strings.EqualFold(debugEnv, "yes")

	debug = true
	if debug {
		// Prepare stdout/stderr pipes for logging only in debug mode
		stdout, err := s.cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %v", err)
		}
		stderr, err := s.cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to get stderr pipe: %v", err)
		}

		logutil.Logf("[ViteServer] Starting (debug): npm %v (dir=%s) URL=%s\n", cmdArgs, projectRoot, s.url)

		// Start the command
		err = s.cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start Vite server: %v", err)
		}

		// Stream logs from Vite process
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				logutil.Logf("[Vite stdout] %s\n", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				logutil.Logf("[Vite stdout error] %v\n", err)
			}
		}()
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				logutil.Logf("[Vite stderr] %s\n", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				logutil.Logf("[Vite stderr error] %v\n", err)
			}
		}()
	} else {
		// Discard stdout/stderr to keep output clean and avoid backpressure
		s.cmd.Stdout = io.Discard
		s.cmd.Stderr = io.Discard
		logutil.Logf("[ViteServer] Starting: npm %v (dir=%s) URL=%s\n", cmdArgs, projectRoot, s.url)
		// Start the command
		err = s.cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start Vite server: %v", err)
		}
	}

	// Wait for the server to be ready
	err = s.waitForServer()
	if err != nil {
		s.Stop()
		return fmt.Errorf("server failed to start: %v", err)
	}

	return nil
}

// Stop stops the Vite development server
func (s *ViteServer) Stop() {
	logutil.Logf("[ViteServer] Stopping: URL=%s\n", s.url)
	debugEnv := os.Getenv("UIWGO_VITE_DEBUG")
	debug := strings.EqualFold(debugEnv, "1") || strings.EqualFold(debugEnv, "true") || strings.EqualFold(debugEnv, "yes")

	if s.cmd != nil && s.cmd.Process != nil {
		pid := s.cmd.Process.Pid

		// On Unix, attempt to gracefully terminate child processes first
		if runtime.GOOS != "windows" {
			// pkill -TERM -P <pid>
			cmd := exec.Command("pkill", "-TERM", "-P", strconv.Itoa(pid))
			_ = cmd.Run()
			// Then signal parent with SIGTERM
			if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
				if debug {
					logutil.Logf("[ViteServer] error sending SIGTERM: %v\n", err)
				}
			}
		}

		// Wait with timeout
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil && debug {
				logutil.Logf("[ViteServer] process exited with error: %v\n", err)
			}
			// Ensure context is canceled after process exit
			if s.cancel != nil {
				s.cancel()
			}
			return
		case <-time.After(5 * time.Second):
			if debug {
				logutil.Logf("[ViteServer] graceful shutdown timed out; forcing kill...\n")
			}
			// Cancel the context to enforce termination
			if s.cancel != nil {
				s.cancel()
			}
			// On Unix, force kill children, then parent
			if runtime.GOOS != "windows" {
				cmd := exec.Command("pkill", "-KILL", "-P", strconv.Itoa(pid))
				_ = cmd.Run()
			}
			_ = s.cmd.Process.Kill()
			select {
			case err := <-done:
				if err != nil && debug {
					logutil.Logf("[ViteServer] process wait error after kill: %v\n", err)
				}
			case <-time.After(3 * time.Second):
				if debug {
					logutil.Logf("[ViteServer] process did not exit after kill timeout\n")
				}
			}
		}
	} else if s.cancel != nil {
		// No process, just cancel context if present
		s.cancel()
	}
}

// URL returns the server URL
func (s *ViteServer) URL() string {
	return s.url
}

// waitForServer waits for the Vite server to be ready
func (s *ViteServer) waitForServer() error {
	// Increased timeout to account for WASM build time
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for server to start")
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.host, s.port), time.Second)
			if err == nil {
				conn.Close()
				// Give it more time for WASM build and initialization
				time.Sleep(5 * time.Second)
				return nil
			}
		}
	}
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (go.mod not found)")
}
