package devserver

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestIndexInjection(t *testing.T) {
	server := NewServer("counter", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	defer server.Stop()

	resp, err := http.Get(server.URL() + "/")
	if err != nil {
		t.Fatalf("failed to GET index: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %v", resp.Status)
	}
	b, _ := io.ReadAll(resp.Body)
	body := string(b)
	if !strings.Contains(body, "/__livereload") || !strings.Contains(body, "EventSource") {
		t.Fatalf("index.html response missing live-reload injection; body: %s", body)
	}
}
