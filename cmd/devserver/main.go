package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const defaultIndexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Go WASM Counter</title>
</head>
<body>
    <h1>Go WASM Counter</h1>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });

        // Use Server-Sent Events for hot reload
        const eventSource = new EventSource('/events');
        eventSource.onmessage = function(event) {
            if (event.data === 'reload') {
                console.log('🔄 Reloading page...');
                window.location.reload();
            }
        };
        
        eventSource.onerror = function(event) {
            console.log('SSE connection error, will reconnect automatically');
        };
    </script>
</body>
</html>`

// SSE connection management
type SSEBroker struct {
	clients    map[chan string]bool
	newClients chan chan string
	closeChan  chan chan string
	messages   chan string
	mutex      sync.RWMutex
}

func NewSSEBroker() *SSEBroker {
	broker := &SSEBroker{
		clients:    make(map[chan string]bool),
		newClients: make(chan chan string),
		closeChan:  make(chan chan string),
		messages:   make(chan string),
	}
	go broker.listen()
	return broker
}

func (broker *SSEBroker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			broker.mutex.Lock()
			broker.clients[s] = true
			broker.mutex.Unlock()
			log.Printf("Client added. %d registered clients", len(broker.clients))

		case s := <-broker.closeChan:
			broker.mutex.Lock()
			delete(broker.clients, s)
			broker.mutex.Unlock()
			log.Printf("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.messages:
			broker.mutex.RLock()
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				default:
					close(clientMessageChan)
					delete(broker.clients, clientMessageChan)
				}
			}
			broker.mutex.RUnlock()
		}
	}
}

func (broker *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan string)
	broker.newClients <- messageChan

	defer func() {
		broker.closeChan <- messageChan
	}()

	notify := r.Context().Done()
	go func() {
		<-notify
		broker.closeChan <- messageChan
	}()

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (broker *SSEBroker) Broadcast(message string) {
	broker.messages <- message
}

var sseBroker *SSEBroker

func main() {
	port := flag.String("port", "8090", "Port to run the server on")
	dir := flag.String("dir", ".", "Directory to serve")
	noWatch := flag.Bool("no-watch", false, "Disable watching for file changes")
	noBuild := flag.Bool("no-build", false, "Disable rebuild at startup")
	flag.Parse()

	// Initialize SSE broker
	sseBroker = NewSSEBroker()

	ensureWasmExec()
	// No longer need to create index.html file since we serve embedded HTML

	if *noBuild && *noWatch {
		log.Println("⚠️ Warning: both --no-build and --no-watch enabled; make sure main.wasm exists")
	}

	if !*noBuild {
		rebuild()
	}

	if !*noWatch {
		go watchAndRebuild()
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// SSE endpoint for hot reload events
	r.Handle("/events", sseBroker)

	// SPA-aware file server that falls back to index.html for routes
	r.Handle("/*", spaHandler(*dir))

	log.Printf("🚀 Golid dev server running at http://localhost:%s (serving from %s)", *port, *dir)
	err := http.ListenAndServe(":"+*port, r)
	if err != nil {
		log.Fatal(err)
	}
}

func watchAndRebuild() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(path, "./cmd") {
				return filepath.SkipDir
			}
			if strings.HasPrefix(path, "./.devenv") {
				return filepath.SkipDir
			}
			log.Println("👀 Adding watch on:", path)
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🚀 Watching for .go file changes (excluding ./cmd)")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && hasGoExtension(event.Name) {
				log.Println("🔨 Change detected in", event.Name, "→ rebuilding WASM...")
				rebuild()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func hasGoExtension(name string) bool {
	return strings.HasSuffix(name, ".go")
}

// spaHandler creates an HTTP handler that serves static files when they exist,
// and falls back to serving index.html for SPA routes
func spaHandler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, r.URL.Path)

		// Check if requesting index.html specifically or a SPA route
		if r.URL.Path == "/" || r.URL.Path == "/index.html" || filepath.Ext(r.URL.Path) == "" {
			// Serve embedded HTML for root, index.html, or SPA routes
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(defaultIndexHTML))
			return
		}

		// Check if the requested path is a file that exists
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			// File exists, serve it normally
			fileServer.ServeHTTP(w, r)
			return
		}

		// File with extension doesn't exist, return 404
		http.NotFound(w, r)
	})
}

func rebuild() {
	cmd := exec.Command("go", "build", "-o", "main.wasm", "./main.go")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	err := cmd.Run()
	if err != nil {
		log.Println("❌ Build failed:", err)
	} else {
		log.Println("✅ Build succeeded")
		// Broadcast reload event to all SSE clients
		if sseBroker != nil {
			sseBroker.Broadcast("reload")
		}
	}
}

func ensureFileExists(filename, defaultContent string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.WriteFile(filename, []byte(defaultContent), 0644)
		if err != nil {
			log.Fatalf("❌ Failed to create %s: %v", filename, err)
		}
		log.Printf("✅ Created %s", filename)
	}
}

// func ensureWasmExec() {
// 	if _, err := os.Stat("wasm_exec.js"); os.IsNotExist(err) {
// 		out, err := exec.Command("go", "env", "GOROOT").Output()
// 		if err != nil {
// 			log.Fatalf("❌ Failed to get GOROOT: %v", err)
// 		}
// 		wasmPath := filepath.Join(strings.TrimSpace(string(out)), "misc", "wasm", "wasm_exec.js")
// 		input, err := os.ReadFile(wasmPath)
// 		if err != nil {
// 			log.Fatalf("❌ Failed to read wasm_exec.js from Go installation: %v", err)
// 		}
// 		err = os.WriteFile("wasm_exec.js", input, 0644)
// 		if err != nil {
// 			log.Fatalf("❌ Failed to write wasm_exec.js to project: %v", err)
// 		}
// 		log.Println("✅ Copied wasm_exec.js from Go installation")
// 	}
// }

func ensureWasmExec() {
	if _, err := os.Stat("wasm_exec.js"); os.IsNotExist(err) {
		out, err := exec.Command("go", "env", "GOROOT").Output()
		if err != nil {
			log.Fatalf("❌ Failed to get GOROOT: %v", err)
		}
		root := strings.TrimSpace(string(out))

		// Try new location first
		candidates := []string{
			filepath.Join(root, "lib", "wasm", "wasm_exec.js"),  // Go 1.21+
			filepath.Join(root, "misc", "wasm", "wasm_exec.js"), // legacy
		}

		var wasmPath string
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				wasmPath = candidate
				break
			}
		}

		if wasmPath == "" {
			log.Fatalf("❌ Could not locate wasm_exec.js in known GOROOT paths.")
		}

		input, err := os.ReadFile(wasmPath)
		if err != nil {
			log.Fatalf("❌ Failed to read wasm_exec.js: %v", err)
		}
		err = os.WriteFile("wasm_exec.js", input, 0644)
		if err != nil {
			log.Fatalf("❌ Failed to write wasm_exec.js: %v", err)
		}
		log.Printf("✅ Copied wasm_exec.js from: %s", wasmPath)
	}
}
