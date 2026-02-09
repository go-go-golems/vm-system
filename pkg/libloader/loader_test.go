package libloader

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

func TestLoadExistingCacheDiscoversBuiltinLibrary(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	lc, err := NewLibraryCache(cacheDir)
	if err != nil {
		t.Fatalf("new library cache: %v", err)
	}

	if err := os.WriteFile(filepath.Join(cacheDir, "lodash-4.17.21.js"), []byte("var _ = {};"), 0o644); err != nil {
		t.Fatalf("write cache fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "custom.js"), []byte("ignored"), 0o644); err != nil {
		t.Fatalf("write non-matching fixture: %v", err)
	}

	if err := lc.LoadExistingCache(); err != nil {
		t.Fatalf("load existing cache: %v", err)
	}

	gotPath, err := lc.GetLibraryPath("lodash")
	if err != nil {
		t.Fatalf("expected lodash to be discovered, got %v", err)
	}
	if filepath.Base(gotPath) != "lodash-4.17.21.js" {
		t.Fatalf("expected discovered filename lodash-4.17.21.js, got %q", filepath.Base(gotPath))
	}

	if _, err := lc.GetLibraryPath("custom"); err == nil {
		t.Fatalf("expected unknown library id to be missing")
	}
}

func TestDownloadCacheHitLoadCodeAndChecksum(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	lc, err := NewLibraryCache(cacheDir)
	if err != nil {
		t.Fatalf("new library cache: %v", err)
	}

	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte("console.log('hello-lib');"))
	}))

	lib := vmmodels.Library{
		ID:      "testlib",
		Name:    "Test Library",
		Version: "1.0.0",
		Source:  server.URL,
	}

	if err := lc.Download(lib); err != nil {
		t.Fatalf("first download: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one HTTP fetch on first download, got %d", hits)
	}

	// Closing the server ensures the second call can only pass if cache hit
	// short-circuits network access.
	server.Close()
	if err := lc.Download(lib); err != nil {
		t.Fatalf("second download should use cache hit, got %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected no extra HTTP hits on cache hit, got %d", hits)
	}

	gotPath, err := lc.GetLibraryPath("testlib")
	if err != nil {
		t.Fatalf("get library path: %v", err)
	}
	if filepath.Base(gotPath) != "testlib-1.0.0.js" {
		t.Fatalf("expected deterministic filename testlib-1.0.0.js, got %q", filepath.Base(gotPath))
	}

	code, err := lc.LoadLibraryCode("testlib")
	if err != nil {
		t.Fatalf("load library code: %v", err)
	}
	if code != "console.log('hello-lib');" {
		t.Fatalf("unexpected library code: %q", code)
	}

	checksum, err := lc.ComputeChecksum("testlib")
	if err != nil {
		t.Fatalf("compute checksum: %v", err)
	}
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte("console.log('hello-lib');")))
	if checksum != expected {
		t.Fatalf("expected checksum %s, got %s", expected, checksum)
	}
}

func TestDownloadReturnsErrorOnUnexpectedHTTPStatus(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	lc, err := NewLibraryCache(cacheDir)
	if err != nil {
		t.Fatalf("new library cache: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	err = lc.Download(vmmodels.Library{
		ID:      "badlib",
		Name:    "Bad Library",
		Version: "9.9.9",
		Source:  server.URL,
	})
	if err == nil {
		t.Fatalf("expected download failure on non-200 response")
	}
	if !strings.Contains(err.Error(), "unexpected status code") {
		t.Fatalf("expected status-code error, got %v", err)
	}
}
