package strategy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/akdavidsson/trawl/internal/config"
)

// cacheKey generates a filename from a URL pattern and fingerprint.
func cacheKey(urlPattern, fingerprint string) string {
	h := sha256.Sum256([]byte(urlPattern + ":" + fingerprint))
	return fmt.Sprintf("%x.json", h[:12])
}

// LoadCached attempts to load a cached strategy matching the URL pattern and fingerprint.
// Returns nil if no cached strategy exists or the fingerprint doesn't match.
func LoadCached(urlPattern, fingerprint string) (*ExtractionStrategy, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, cacheKey(urlPattern, fingerprint))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no cached strategy
		}
		return nil, fmt.Errorf("reading cached strategy: %w", err)
	}

	var s ExtractionStrategy
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing cached strategy: %w", err)
	}

	// Verify fingerprint matches
	if s.Fingerprint != fingerprint {
		return nil, nil // fingerprint mismatch, need re-derivation
	}

	return &s, nil
}

// SaveCache writes a strategy to the cache directory.
func SaveCache(s *ExtractionStrategy) error {
	dir, err := config.CacheDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling strategy: %w", err)
	}

	path := filepath.Join(dir, cacheKey(s.SitePattern, s.Fingerprint))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing cached strategy: %w", err)
	}

	return nil
}

// LoadFromFile loads a strategy from a specific file path.
func LoadFromFile(path string) (*ExtractionStrategy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading strategy file: %w", err)
	}

	var s ExtractionStrategy
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing strategy file: %w", err)
	}

	return &s, nil
}
