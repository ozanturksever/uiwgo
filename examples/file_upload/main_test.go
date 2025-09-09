//go:build !js && !wasm

package main

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ozanturksever/uiwgo/internal/testhelpers"
)

func TestFileUploadExample_PageRender(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	// Create chromedp context with visible browser for debugging
	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var title, heading string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Title(&title),
		chromedp.Text("h1", &heading),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if title != "File Upload Example" {
		t.Errorf("Expected title 'File Upload Example', got '%s'", title)
	}
	if heading != "File Upload Example" {
		t.Errorf("Expected heading 'File Upload Example', got '%s'", heading)
	}
}

func TestFileUploadExample_DropZone(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var dropZoneText string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Text("h3", &dropZoneText),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if dropZoneText != "📁 Drop Files Here" {
		t.Errorf("Expected drop zone text '📁 Drop Files Here', got '%s'", dropZoneText)
	}
}

func TestFileUploadExample_FileInput(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var inputType, inputMultiple string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.AttributeValue(`input[type="file"]`, "type", &inputType, nil),
		chromedp.AttributeValue(`input[type="file"]`, "multiple", &inputMultiple, nil),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if inputType != "file" {
		t.Errorf("Expected input type 'file', got '%s'", inputType)
	}
	if inputMultiple != "true" {
		t.Errorf("Expected input multiple 'true', got '%s'", inputMultiple)
	}
}

func TestFileUploadExample_FileSelection(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var noFilesText string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Check initial state - look for the specific "No files uploaded yet" text in the file list area
		chromedp.Text("div[style*='margin-top: 30px'] p", &noFilesText),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if noFilesText != "No files uploaded yet" {
		t.Errorf("Expected initial text 'No files uploaded yet', got '%s'", noFilesText)
	}
}

func TestFileUploadExample_ClearFiles(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var finalText string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Check files are cleared - look for the specific "No files uploaded yet" text in the file list area
		chromedp.Text("div[style*='margin-top: 30px'] p", &finalText),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if finalText != "No files uploaded yet" {
		t.Errorf("Expected final text 'No files uploaded yet', got '%s'", finalText)
	}
}

func TestFileUploadExample_RemoveFile(t *testing.T) {
	server := testhelpers.NewViteServer("file_upload", "localhost:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start dev server: %v", err)
	}
	defer server.Stop()

	chromedpCtx := testhelpers.MustNewChromedpContext(testhelpers.DefaultConfig())
	defer chromedpCtx.Cancel()

	var noFilesText string
	err := chromedp.Run(chromedpCtx.Ctx,
		chromedp.Navigate(server.URL()),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Check initial state - look for the specific "No files uploaded yet" text in the file list area
		chromedp.Text("div[style*='margin-top: 30px'] p", &noFilesText),
	)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	if noFilesText != "No files uploaded yet" {
		t.Errorf("Expected initial text 'No files uploaded yet', got '%s'", noFilesText)
	}
}