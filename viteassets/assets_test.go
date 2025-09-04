package viteassets

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDefaultDevConfig tests the default development configuration
func TestDefaultDevConfig(t *testing.T) {
	baseURL := "http://localhost:5173"
	cfg := DefaultDevConfig(baseURL)
	
	if cfg.BaseURL != baseURL {
		t.Errorf("Expected BaseURL to be %s, got %s", baseURL, cfg.BaseURL)
	}
	
	if !cfg.DevMode {
		t.Error("Expected DevMode to be true")
	}
	
	if cfg.ManifestPath != "" {
		t.Errorf("Expected ManifestPath to be empty, got %s", cfg.ManifestPath)
	}
	
	if cfg.ManifestURL != "" {
		t.Errorf("Expected ManifestURL to be empty, got %s", cfg.ManifestURL)
	}
}

// TestDefaultProdConfig tests the default production configuration
func TestDefaultProdConfig(t *testing.T) {
	baseURL := "https://example.com"
	manifestURL := "https://example.com/manifest.json"
	cfg := DefaultProdConfig(baseURL, manifestURL)
	
	if cfg.BaseURL != baseURL {
		t.Errorf("Expected BaseURL to be %s, got %s", baseURL, cfg.BaseURL)
	}
	
	if cfg.DevMode {
		t.Error("Expected DevMode to be false")
	}
	
	if cfg.ManifestPath != "" {
		t.Errorf("Expected ManifestPath to be empty, got %s", cfg.ManifestPath)
	}
	
	if cfg.ManifestURL != manifestURL {
		t.Errorf("Expected ManifestURL to be %s, got %s", manifestURL, cfg.ManifestURL)
	}
}

// TestNewAssetResolver tests creating a new asset resolver
func TestNewAssetResolver(t *testing.T) {
	cfg := DefaultDevConfig("http://localhost:5173")
	resolver := NewAssetResolver(cfg)
	
	if resolver == nil {
		t.Error("Expected resolver to be created")
	}
	
	if resolver.GetBaseURL() != cfg.BaseURL {
		t.Errorf("Expected base URL to be %s, got %s", cfg.BaseURL, resolver.GetBaseURL())
	}
	
	if !resolver.IsDevMode() {
		t.Error("Expected resolver to be in dev mode")
	}
}

// TestResolveAssetInDevMode tests asset resolution in development mode
func TestResolveAssetInDevMode(t *testing.T) {
	baseURL := "http://localhost:5173"
	cfg := DefaultDevConfig(baseURL)
	resolver := NewAssetResolver(cfg)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"main.js", "http://localhost:5173/main.js"},
		{"/main.js", "http://localhost:5173/main.js"},
		{"assets/style.css", "http://localhost:5173/assets/style.css"},
		{"/assets/style.css", "http://localhost:5173/assets/style.css"},
	}
	
	for _, test := range tests {
		result := resolver.ResolveAsset(test.input)
		if result != test.expected {
			t.Errorf("For input %s, expected %s, got %s", test.input, test.expected, result)
		}
	}
}

// TestLoadManifestInDevMode tests that manifest loading is skipped in dev mode
func TestLoadManifestInDevMode(t *testing.T) {
	cfg := DefaultDevConfig("http://localhost:5173")
	resolver := NewAssetResolver(cfg)
	
	err := resolver.LoadManifest()
	if err != nil {
		t.Errorf("Expected no error in dev mode, got %v", err)
	}
}

// TestLoadManifestWithoutURL tests manifest loading without URL
func TestLoadManifestWithoutURL(t *testing.T) {
	cfg := DefaultProdConfig("https://example.com", "")
	resolver := NewAssetResolver(cfg)
	
	err := resolver.LoadManifest()
	if err == nil {
		t.Error("Expected error when manifest URL is not configured")
	}
}

// TestLoadManifestSuccess tests successful manifest loading
func TestLoadManifestSuccess(t *testing.T) {
	// Create a test manifest
	manifest := map[string]ManifestEntry{
		"main.js": {
			File:    "assets/main-abc123.js",
			Src:     "main.js",
			IsEntry: true,
			CSS:     []string{"assets/main-def456.css"},
		},
		"style.css": {
			File: "assets/style-ghi789.css",
			Src:  "style.css",
		},
	}
	
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manifest)
	}))
	defer server.Close()
	
	// Create resolver with test server URL
	cfg := DefaultProdConfig("https://example.com", server.URL)
	resolver := NewAssetResolver(cfg)
	
	// Load manifest
	err := resolver.LoadManifest()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test asset resolution
	result := resolver.ResolveAsset("main.js")
	expected := "https://example.com/assets/main-abc123.js"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestLoadManifestHTTPError tests manifest loading with HTTP error
func TestLoadManifestHTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	
	// Create resolver with test server URL
	cfg := DefaultProdConfig("https://example.com", server.URL)
	resolver := NewAssetResolver(cfg)
	
	// Load manifest should fail
	err := resolver.LoadManifest()
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

// TestLoadManifestInvalidJSON tests manifest loading with invalid JSON
func TestLoadManifestInvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()
	
	// Create resolver with test server URL
	cfg := DefaultProdConfig("https://example.com", server.URL)
	resolver := NewAssetResolver(cfg)
	
	// Load manifest should fail
	err := resolver.LoadManifest()
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestGetEntryAssetsInDevMode tests getting entry assets in dev mode
func TestGetEntryAssetsInDevMode(t *testing.T) {
	cfg := DefaultDevConfig("http://localhost:5173")
	resolver := NewAssetResolver(cfg)
	
	assets := resolver.GetEntryAssets("main.js")
	if len(assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(assets))
	}
	
	expected := "http://localhost:5173/main.js"
	if assets[0] != expected {
		t.Errorf("Expected %s, got %s", expected, assets[0])
	}
}

// TestGetCSSAssetsInDevMode tests getting CSS assets in dev mode
func TestGetCSSAssetsInDevMode(t *testing.T) {
	cfg := DefaultDevConfig("http://localhost:5173")
	resolver := NewAssetResolver(cfg)
	
	assets := resolver.GetCSSAssets("main.js")
	if len(assets) != 0 {
		t.Errorf("Expected 0 CSS assets in dev mode, got %d", len(assets))
	}
}

// TestGlobalResolver tests the global resolver functionality
func TestGlobalResolver(t *testing.T) {
	// Initially should be nil
	if GetGlobalResolver() != nil {
		t.Error("Expected global resolver to be nil initially")
	}
	
	// Create and set a resolver
	cfg := DefaultDevConfig("http://localhost:5173")
	resolver := NewAssetResolver(cfg)
	SetGlobalResolver(resolver)
	
	// Should now return the set resolver
	if GetGlobalResolver() != resolver {
		t.Error("Expected GetGlobalResolver to return the set resolver")
	}
	
	// Test global convenience functions
	asset := Asset("main.js")
	expected := "http://localhost:5173/main.js"
	if asset != expected {
		t.Errorf("Expected %s, got %s", expected, asset)
	}
	
	assets := EntryAssets("main.js")
	if len(assets) != 1 || assets[0] != expected {
		t.Errorf("Expected [%s], got %v", expected, assets)
	}
	
	cssAssets := CSSAssets("main.js")
	if len(cssAssets) != 0 {
		t.Errorf("Expected 0 CSS assets, got %d", len(cssAssets))
	}
	
	// Clear global resolver
	SetGlobalResolver(nil)
	if GetGlobalResolver() != nil {
		t.Error("Expected global resolver to be nil after clearing")
	}
}

// TestGlobalFunctionsWithoutResolver tests global functions without a resolver
func TestGlobalFunctionsWithoutResolver(t *testing.T) {
	// Ensure no global resolver is set
	SetGlobalResolver(nil)
	
	// Asset should return the path as-is
	asset := Asset("main.js")
	if asset != "main.js" {
		t.Errorf("Expected 'main.js', got %s", asset)
	}
	
	// EntryAssets should return the entry path only
	assets := EntryAssets("main.js")
	if len(assets) != 1 || assets[0] != "main.js" {
		t.Errorf("Expected ['main.js'], got %v", assets)
	}
	
	// CSSAssets should return empty slice
	cssAssets := CSSAssets("main.js")
	if len(cssAssets) != 0 {
		t.Errorf("Expected 0 CSS assets, got %d", len(cssAssets))
	}
}

// TestInitDev tests the InitDev convenience function
func TestInitDev(t *testing.T) {
	baseURL := "http://localhost:5173"
	InitDev(baseURL)
	
	resolver := GetGlobalResolver()
	if resolver == nil {
		t.Error("Expected global resolver to be set after InitDev")
	}
	
	if !resolver.IsDevMode() {
		t.Error("Expected resolver to be in dev mode")
	}
	
	if resolver.GetBaseURL() != baseURL {
		t.Errorf("Expected base URL to be %s, got %s", baseURL, resolver.GetBaseURL())
	}
}

// TestInitProd tests the InitProd convenience function
func TestInitProd(t *testing.T) {
	// Create a test manifest
	manifest := map[string]ManifestEntry{
		"main.js": {
			File:    "assets/main-abc123.js",
			Src:     "main.js",
			IsEntry: true,
		},
	}
	
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manifest)
	}))
	defer server.Close()
	
	baseURL := "https://example.com"
	err := InitProd(baseURL, server.URL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	resolver := GetGlobalResolver()
	if resolver == nil {
		t.Error("Expected global resolver to be set after InitProd")
	}
	
	if resolver.IsDevMode() {
		t.Error("Expected resolver to be in prod mode")
	}
	
	if resolver.GetBaseURL() != baseURL {
		t.Errorf("Expected base URL to be %s, got %s", baseURL, resolver.GetBaseURL())
	}
}

// TestReloadManifest tests the ReloadManifest convenience function
func TestReloadManifest(t *testing.T) {
	// Without global resolver should return error
	SetGlobalResolver(nil)
	err := ReloadManifest()
	if err == nil {
		t.Error("Expected error when no global resolver is configured")
	}
	
	// With global resolver in dev mode should work
	InitDev("http://localhost:5173")
	err = ReloadManifest()
	if err != nil {
		t.Errorf("Expected no error in dev mode, got %v", err)
	}
}

// TestManifestEntry tests the ManifestEntry struct
func TestManifestEntry(t *testing.T) {
	entry := ManifestEntry{
		File:    "assets/main-abc123.js",
		Src:     "main.js",
		IsEntry: true,
		CSS:     []string{"assets/main-def456.css"},
		Assets:  []string{"assets/logo-ghi789.png"},
	}
	
	if entry.File != "assets/main-abc123.js" {
		t.Errorf("Expected File to be 'assets/main-abc123.js', got %s", entry.File)
	}
	
	if entry.Src != "main.js" {
		t.Errorf("Expected Src to be 'main.js', got %s", entry.Src)
	}
	
	if !entry.IsEntry {
		t.Error("Expected IsEntry to be true")
	}
	
	if len(entry.CSS) != 1 || entry.CSS[0] != "assets/main-def456.css" {
		t.Errorf("Expected CSS to be ['assets/main-def456.css'], got %v", entry.CSS)
	}
	
	if len(entry.Assets) != 1 || entry.Assets[0] != "assets/logo-ghi789.png" {
		t.Errorf("Expected Assets to be ['assets/logo-ghi789.png'], got %v", entry.Assets)
	}
}