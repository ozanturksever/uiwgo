package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Simple SSE hub
type sseHub struct {
	clients map[chan string]struct{}
	mu      sync.Mutex
}

func newSSEHub() *sseHub { return &sseHub{clients: make(map[chan string]struct{})} }

func (h *sseHub) addClient(ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[ch] = struct{}{}
}

func (h *sseHub) removeClient(ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ch)
	close(ch)
}

func (h *sseHub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func serveWithLiveReload(hub *sseHub, example string) {
	// Static files from examples/<example>
	dir := filepath.Join("examples", example)
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	// wasm_exec.js from repo root
	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(".", "wasm_exec.js"))
	})

	// SSE endpoint
	http.HandleFunc("/__livereload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		ch := make(chan string, 8)
		hub.addClient(ch)
		defer hub.removeClient(ch)
		// Send a ping to establish connection
		fmt.Fprintf(w, "event: ping\n")
		fmt.Fprintf(w, "data: ok\n\n")
		flusher.Flush()
		// Stream messages
		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	})
}

func buildWASM(example string) error {
	log.Printf("==> Building WASM binary for '%s' example...\n", example)
	outPath := filepath.Join("examples", example, "main.wasm")
	srcPath := filepath.Join("examples", example, "main.go")
	cmd := exec.Command("go", "build", "-o", outPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
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

func watchAndReload(ctx context.Context, hub *sseHub, example string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	exampleDir := filepath.Join("examples", example)
	paths := []string{
		exampleDir,
		"pkg",
		"server.go",
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					// Skip hidden dirs like .git
					if strings.HasPrefix(info.Name(), ".") {
						return filepath.SkipDir
					}
					_ = watcher.Add(path)
				}
				return nil
			})
		} else {
			_ = watcher.Add(p)
		}
	}

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}

	rebuild := func() {
		if err := buildWASM(example); err != nil {
			log.Println("[dev] Build failed:", err)
			return
		}
		hub.broadcast("reload")
		log.Println("[dev] Reload signaled")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev := <-watcher.Events:
			// Consider only relevant extensions
			name := strings.ToLower(ev.Name)
			if !(strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".html")) {
				continue
			}
			// Debounce rapid events
			debounce.Reset(200 * time.Millisecond)
		case <-debounce.C:
			rebuild()
		case err := <-watcher.Errors:
			log.Println("[dev] watcher error:", err)
		}
	}
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	example := flag.String("example", "counter", "example directory under ./examples to run")
	flag.Parse()

	exampleDir := filepath.Join("examples", *example)
	if info, err := os.Stat(exampleDir); err != nil || !info.IsDir() {
		log.Fatalf("example '%s' not found at %s", *example, exampleDir)
	}

	if _, err := os.Stat(filepath.Join(exampleDir, "main.go")); err != nil {
		log.Fatalf("example '%s' missing main.go", *example)
	}

	if err := buildWASM(*example); err != nil {
		log.Println("Initial build failed:", err)
	}

	hub := newSSEHub()
	serveWithLiveReload(hub, *example)

	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Printf("==> Serving http://localhost:8080 (example: %s)\n", *example)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go func() {
		if err := watchAndReload(ctx, hub, *example); err != nil {
			log.Println("watch error:", err)
		}
	}()

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	// Use standard signals
	// Imported lazily to keep imports tidy
	//goland:noinspection GoDeprecation
	signalNotify(sigCh)
	<-sigCh
	log.Println("Shutting down...")
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// small wrapper to register for SIGINT/SIGTERM
func signalNotify(ch chan os.Signal) {
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}
