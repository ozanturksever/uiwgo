package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ozanturksever/uiwgo/internal/devserver"
)

// Simple SSE hub for live reload
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

func addLiveReloadEndpoint(server *devserver.Server, hub *sseHub) {
	// SSE endpoint for live reload
	server.HandleFunc("/__livereload", func(w http.ResponseWriter, r *http.Request) {
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
		"comps",
		"dom",
		"reactivity",
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
		if err := devserver.BuildWASM(example); err != nil {
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
	port := flag.Int("port", 8080, "port to serve the dev server on")
	flag.Parse()

	exampleDir := filepath.Join("examples", *example)
	if info, err := os.Stat(exampleDir); err != nil || !info.IsDir() {
		log.Fatalf("example '%s' not found at %s", *example, exampleDir)
	}

	if _, err := os.Stat(filepath.Join(exampleDir, "main.go")); err != nil {
		log.Fatalf("example '%s' missing main.go", *example)
	}

	// Create and start the development server
	server := devserver.NewServer(*example, fmt.Sprintf(":%d", *port))
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Add live reload endpoint
	hub := newSSEHub()
	addLiveReloadEndpoint(server, hub)

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
	server.Stop()
}

// small wrapper to register for SIGINT/SIGTERM
func signalNotify(ch chan os.Signal) {
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}
