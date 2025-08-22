package devserver

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "embed"
)

//go:embed wasm_exec.js
var wasmExecJS []byte



// BuildWASM compiles the Go code to WebAssembly for the given example
func BuildWASM(example string) error {
	log.Printf("==> Building WASM binary for '%s' example...\n", example)
	
	// Determine the correct paths based on current working directory
	var outPath, srcPath, workDir string
	
	// Check if we're already in the example directory
	if _, err := os.Stat("main.go"); err == nil {
		// We're in the example directory
		outPath = "main.wasm"
		srcPath = "main.go"
		workDir = "."
	} else {
		// We're in the project root or elsewhere
		outPath = filepath.Join("examples", example, "main.wasm")
		srcPath = filepath.Join("examples", example, "main.go")
		workDir = "."
		
		// Check if examples directory exists
		if _, err := os.Stat(filepath.Join("examples", example)); err != nil {
			// Try going up one level
			outPath = filepath.Join("..", "..", "examples", example, "main.wasm")
			srcPath = filepath.Join("..", "..", "examples", example, "main.go")
			workDir = filepath.Join("..", "..")
		}
	}
	
	cmd := exec.Command("go", "build", "-o", outPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		s := string(out)
		// Print only interesting lines
		scanner := bufio.NewScanner(strings.NewReader(s))
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}
	return err
}

// Server represents a development server instance
type Server struct {
	server   *http.Server
	example  string
	addr     string
	listener net.Listener
}

// NewServer creates a new development server for the given example
// If addr is empty or "localhost:0", it will use a random available port
func NewServer(example, addr string) *Server {
	if addr == "" || addr == "localhost:0" {
		addr = "localhost:0"
	}
	return &Server{
		example: example,
		addr:    addr,
	}
}

// Start starts the development server
func (s *Server) Start() error {
	// Build WASM first
	if err := BuildWASM(s.example); err != nil {
		return fmt.Errorf("failed to build WASM: %w", err)
	}

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Static files from examples/<example> or current directory
	var dir string
	if _, err := os.Stat("index.html"); err == nil {
		// We're in the example directory
		dir = "."
	} else {
		// We're in the project root or elsewhere
		dir = filepath.Join("examples", s.example)
		// Check if examples directory exists
		if _, err := os.Stat(dir); err != nil {
			// Try going up levels
			dir = filepath.Join("..", "..", "examples", s.example)
		}
	}
	fs := http.FileServer(http.Dir(dir))
	mux.Handle("/", fs)

	// wasm_exec.js served from embedded content
	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(wasmExecJS)
	})

	// Create listener to get actual port
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener
	s.addr = listener.Addr().String()

	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		log.Printf("==> Serving http://%s (example: %s)\n", s.addr, s.example)
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Stop stops the development server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if s.listener != nil {
		s.listener.Close()
	}
	return err
}

// URL returns the server's base URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s", s.addr)
}