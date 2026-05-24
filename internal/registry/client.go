package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Fetch downloads and parses the registry from the given URL via HTTP GET.
// It uses a new http.Client per call (no global mutable state).
// The context is used for cancellation; returns a wrapped error on non-200
// or network failure.
func Fetch(ctx context.Context, url string) (*Catalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating fetch request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing fetch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var reg Catalog
	if err := json.Unmarshal(body, &reg); err != nil {
		return nil, fmt.Errorf("parsing registry JSON: %w", err)
	}

	return &reg, nil
}

// LoadCache reads and parses a Catalog from a local JSON file path.
// Returns an error wrapping the cause on read or parse failure.
func LoadCache(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	var reg Catalog
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parsing cache JSON: %w", err)
	}

	return &reg, nil
}

// SaveCache writes a Catalog as JSON to a local file path with 0644 permissions.
func SaveCache(path string, r *Catalog) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// IsStale returns true if the cache file does not exist or is older than maxAge.
func IsStale(path string, maxAge time.Duration) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	return time.Since(info.ModTime()) > maxAge
}
