package libloader

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LibraryCache manages downloaded library files
type LibraryCache struct {
	cacheDir string
	mu       sync.RWMutex
	cached   map[string]string // library ID -> local file path
	logger   zerolog.Logger
}

// NewLibraryCache creates a new library cache
func NewLibraryCache(cacheDir string) (*LibraryCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &LibraryCache{
		cacheDir: cacheDir,
		cached:   make(map[string]string),
		logger:   log.With().Str("component", "library_cache").Logger(),
	}, nil
}

// DownloadAll downloads all builtin libraries upfront
func (lc *LibraryCache) DownloadAll() error {
	libraries := vmmodels.BuiltinLibraries()

	lc.logger.Info().
		Int("library_count", len(libraries)).
		Msg("downloading builtin libraries")

	var wg sync.WaitGroup
	errChan := make(chan error, len(libraries))

	for _, lib := range libraries {
		wg.Add(1)
		go func(library vmmodels.Library) {
			defer wg.Done()

			if err := lc.Download(library); err != nil {
				errChan <- fmt.Errorf("failed to download %s: %w", library.Name, err)
			} else {
				lc.logger.Info().
					Str("library", library.ID).
					Str("name", library.Name).
					Str("version", library.Version).
					Msg("downloaded library")
			}
		}(lib)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to download %d libraries: %v", len(errors), errors[0])
	}

	lc.logger.Info().Msg("all builtin libraries downloaded successfully")
	return nil
}

// Download downloads a single library and caches it
func (lc *LibraryCache) Download(lib vmmodels.Library) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Generate cache filename based on library ID and version
	filename := fmt.Sprintf("%s-%s.js", lib.ID, lib.Version)
	cachePath := filepath.Join(lc.cacheDir, filename)

	// Check if already cached
	if _, err := os.Stat(cachePath); err == nil {
		lc.cached[lib.ID] = cachePath
		return nil
	}

	// Download from source
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(lib.Source)
	if err != nil {
		return fmt.Errorf("failed to fetch library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create temporary file
	tmpPath := cachePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Copy content
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write library: %w", err)
	}

	// Rename to final location
	if err := os.Rename(tmpPath, cachePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move library: %w", err)
	}

	lc.cached[lib.ID] = cachePath
	return nil
}

// GetLibraryPath returns the local path for a cached library
func (lc *LibraryCache) GetLibraryPath(libraryID string) (string, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	path, ok := lc.cached[libraryID]
	if !ok {
		return "", fmt.Errorf("library %s not cached", libraryID)
	}

	return path, nil
}

// LoadLibraryCode reads the library code from cache
func (lc *LibraryCache) LoadLibraryCode(libraryID string) (string, error) {
	path, err := lc.GetLibraryPath(libraryID)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read library: %w", err)
	}

	return string(data), nil
}

// LoadExistingCache scans the cache directory and loads existing libraries
func (lc *LibraryCache) LoadExistingCache() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	entries, err := os.ReadDir(lc.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	libraries := vmmodels.BuiltinLibraries()
	libraryMap := make(map[string]vmmodels.Library)
	for _, lib := range libraries {
		libraryMap[lib.ID] = lib
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".js" {
			continue
		}

		// Try to match filename to library
		for id, lib := range libraryMap {
			expectedName := fmt.Sprintf("%s-%s.js", lib.ID, lib.Version)
			if entry.Name() == expectedName {
				lc.cached[id] = filepath.Join(lc.cacheDir, entry.Name())
				break
			}
		}
	}

	return nil
}

// GetCacheInfo returns information about cached libraries
func (lc *LibraryCache) GetCacheInfo() map[string]CacheInfo {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	info := make(map[string]CacheInfo)

	for id, path := range lc.cached {
		stat, err := os.Stat(path)
		if err != nil {
			continue
		}

		info[id] = CacheInfo{
			Path:         path,
			Size:         stat.Size(),
			ModifiedTime: stat.ModTime(),
		}
	}

	return info
}

// CacheInfo contains information about a cached library
type CacheInfo struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
}

// ComputeChecksum computes SHA256 checksum of a cached library
func (lc *LibraryCache) ComputeChecksum(libraryID string) (string, error) {
	path, err := lc.GetLibraryPath(libraryID)
	if err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
