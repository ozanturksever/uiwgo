package viteassets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/ozanturksever/logutil"
)

// AssetResolver handles asset path resolution for Vite dev/prod environments
type AssetResolver struct {
	mu          sync.RWMutex
	manifest    map[string]ManifestEntry
	baseURL     string
	devMode     bool
	manifestURL string
}

// ManifestEntry represents an entry in Vite's manifest.json
type ManifestEntry struct {
	File    string   `json:"file"`
	Src     string   `json:"src,omitempty"`
	IsEntry bool     `json:"isEntry,omitempty"`
	CSS     []string `json:"css,omitempty"`
	Assets  []string `json:"assets,omitempty"`
}

// Config holds configuration for the asset resolver
type Config struct {
	// BaseURL is the base URL for assets (e.g., "http://localhost:5173" for dev)
	BaseURL string
	// DevMode indicates if we're in development mode
	DevMode bool
	// ManifestPath is the path to the manifest.json file (for production)
	ManifestPath string
	// ManifestURL is the URL to fetch manifest.json from (for production)
	ManifestURL string
}

// DefaultDevConfig returns a default configuration for development
func DefaultDevConfig(baseURL string) Config {
	return Config{
		BaseURL:      baseURL,
		DevMode:      true,
		ManifestPath: "",
		ManifestURL:  "",
	}
}

// DefaultProdConfig returns a default configuration for production
func DefaultProdConfig(baseURL, manifestURL string) Config {
	return Config{
		BaseURL:      baseURL,
		DevMode:      false,
		ManifestPath: "",
		ManifestURL:  manifestURL,
	}
}

// NewAssetResolver creates a new asset resolver
func NewAssetResolver(cfg Config) *AssetResolver {
	return &AssetResolver{
		baseURL:     cfg.BaseURL,
		devMode:     cfg.DevMode,
		manifestURL: cfg.ManifestURL,
		manifest:    make(map[string]ManifestEntry),
	}
}

// LoadManifest loads the Vite manifest for production builds
func (ar *AssetResolver) LoadManifest() error {
	if ar.devMode {
		logutil.Log("Skipping manifest load in dev mode")
		return nil
	}

	if ar.manifestURL == "" {
		return fmt.Errorf("manifest URL not configured")
	}

	logutil.Logf("Loading manifest from %s", ar.manifestURL)

	resp, err := http.Get(ar.manifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("manifest request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read manifest body: %w", err)
	}

	var manifest map[string]ManifestEntry
	if err := json.Unmarshal(body, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	ar.mu.Lock()
	ar.manifest = manifest
	ar.mu.Unlock()

	logutil.Logf("Loaded manifest with %d entries", len(manifest))
	return nil
}

// ResolveAsset resolves an asset path based on dev/prod mode
func (ar *AssetResolver) ResolveAsset(assetPath string) string {
	if ar.devMode {
		return ar.resolveDevAsset(assetPath)
	}
	return ar.resolveProdAsset(assetPath)
}

// resolveDevAsset resolves asset paths for development mode
func (ar *AssetResolver) resolveDevAsset(assetPath string) string {
	// In dev mode, assets are served directly by Vite dev server
	cleanPath := strings.TrimPrefix(assetPath, "/")
	return ar.baseURL + "/" + cleanPath
}

// resolveProdAsset resolves asset paths for production mode using manifest
func (ar *AssetResolver) resolveProdAsset(assetPath string) string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	// Clean the asset path
	cleanPath := strings.TrimPrefix(assetPath, "/")

	// Look up in manifest
	if entry, exists := ar.manifest[cleanPath]; exists {
		return ar.baseURL + "/" + entry.File
	}

	// Fallback: try direct path
	logutil.Logf("Asset not found in manifest, using direct path: %s", cleanPath)
	return ar.baseURL + "/" + cleanPath
}

// GetEntryAssets returns all assets for a given entry point
func (ar *AssetResolver) GetEntryAssets(entryPath string) []string {
	if ar.devMode {
		// In dev mode, just return the entry itself
		return []string{ar.resolveDevAsset(entryPath)}
	}

	ar.mu.RLock()
	defer ar.mu.RUnlock()

	cleanPath := strings.TrimPrefix(entryPath, "/")
	entry, exists := ar.manifest[cleanPath]
	if !exists {
		return []string{ar.resolveProdAsset(entryPath)}
	}

	assets := []string{ar.baseURL + "/" + entry.File}

	// Add CSS files
	for _, css := range entry.CSS {
		assets = append(assets, ar.baseURL+"/"+css)
	}

	// Add other assets
	for _, asset := range entry.Assets {
		assets = append(assets, ar.baseURL+"/"+asset)
	}

	return assets
}

// GetCSSAssets returns CSS assets for a given entry point
func (ar *AssetResolver) GetCSSAssets(entryPath string) []string {
	if ar.devMode {
		// In dev mode, CSS is typically injected by Vite
		return []string{}
	}

	ar.mu.RLock()
	defer ar.mu.RUnlock()

	cleanPath := strings.TrimPrefix(entryPath, "/")
	entry, exists := ar.manifest[cleanPath]
	if !exists {
		return []string{}
	}

	var cssAssets []string
	for _, css := range entry.CSS {
		cssAssets = append(cssAssets, ar.baseURL+"/"+css)
	}

	return cssAssets
}

// IsDevMode returns true if the resolver is in development mode
func (ar *AssetResolver) IsDevMode() bool {
	return ar.devMode
}

// GetBaseURL returns the base URL for assets
func (ar *AssetResolver) GetBaseURL() string {
	return ar.baseURL
}

// Global resolver instance
var (
	globalResolver *AssetResolver
	globalMu       sync.RWMutex
)

// SetGlobalResolver sets the global asset resolver
func SetGlobalResolver(resolver *AssetResolver) {
	globalMu.Lock()
	globalResolver = resolver
	globalMu.Unlock()
}

// GetGlobalResolver returns the global asset resolver
func GetGlobalResolver() *AssetResolver {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalResolver
}

// Asset resolves an asset path using the global resolver
func Asset(assetPath string) string {
	resolver := GetGlobalResolver()
	if resolver == nil {
		logutil.Log("No global asset resolver configured, returning path as-is")
		return assetPath
	}
	return resolver.ResolveAsset(assetPath)
}

// EntryAssets returns all assets for an entry using the global resolver
func EntryAssets(entryPath string) []string {
	resolver := GetGlobalResolver()
	if resolver == nil {
		logutil.Log("No global asset resolver configured, returning entry path only")
		return []string{entryPath}
	}
	return resolver.GetEntryAssets(entryPath)
}

// CSSAssets returns CSS assets for an entry using the global resolver
func CSSAssets(entryPath string) []string {
	resolver := GetGlobalResolver()
	if resolver == nil {
		return []string{}
	}
	return resolver.GetCSSAssets(entryPath)
}

// InitDev initializes the global resolver for development
func InitDev(baseURL string) {
	resolver := NewAssetResolver(DefaultDevConfig(baseURL))
	SetGlobalResolver(resolver)
	logutil.Logf("Initialized asset resolver for development with base URL: %s", baseURL)
}

// InitProd initializes the global resolver for production
func InitProd(baseURL, manifestURL string) error {
	resolver := NewAssetResolver(DefaultProdConfig(baseURL, manifestURL))
	if err := resolver.LoadManifest(); err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}
	SetGlobalResolver(resolver)
	logutil.Logf("Initialized asset resolver for production with base URL: %s", baseURL)
	return nil
}

// ReloadManifest reloads the manifest for the global resolver
func ReloadManifest() error {
	resolver := GetGlobalResolver()
	if resolver == nil {
		return fmt.Errorf("no global resolver configured")
	}
	return resolver.LoadManifest()
}
