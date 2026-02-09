package vmdaemon

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfigValues(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig("/tmp/vm-system.db")
	if cfg.DBPath != "/tmp/vm-system.db" {
		t.Fatalf("expected db path to round-trip, got %q", cfg.DBPath)
	}
	if cfg.ListenAddr != "127.0.0.1:3210" {
		t.Fatalf("expected default listen addr 127.0.0.1:3210, got %q", cfg.ListenAddr)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Fatalf("expected read timeout 15s, got %s", cfg.ReadTimeout)
	}
	if cfg.ReadHeaderTime != 5*time.Second {
		t.Fatalf("expected read header timeout 5s, got %s", cfg.ReadHeaderTime)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Fatalf("expected write timeout 30s, got %s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Fatalf("expected idle timeout 60s, got %s", cfg.IdleTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected shutdown timeout 10s, got %s", cfg.ShutdownTimeout)
	}
}

func TestNewConfiguresHTTPServerFromConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig(filepath.Join(t.TempDir(), "vm-system.db"))
	cfg.ListenAddr = "127.0.0.1:0"
	cfg.ReadTimeout = 3 * time.Second
	cfg.ReadHeaderTime = 2 * time.Second
	cfg.WriteTimeout = 4 * time.Second
	cfg.IdleTimeout = 5 * time.Second

	handler := http.NewServeMux()
	app, err := New(cfg, handler)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	defer app.Close()

	if app.server.Addr != cfg.ListenAddr {
		t.Fatalf("expected server addr %q, got %q", cfg.ListenAddr, app.server.Addr)
	}
	if app.server.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("expected read timeout %s, got %s", cfg.ReadTimeout, app.server.ReadTimeout)
	}
	if app.server.ReadHeaderTimeout != cfg.ReadHeaderTime {
		t.Fatalf("expected read header timeout %s, got %s", cfg.ReadHeaderTime, app.server.ReadHeaderTimeout)
	}
	if app.server.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("expected write timeout %s, got %s", cfg.WriteTimeout, app.server.WriteTimeout)
	}
	if app.server.IdleTimeout != cfg.IdleTimeout {
		t.Fatalf("expected idle timeout %s, got %s", cfg.IdleTimeout, app.server.IdleTimeout)
	}
	if app.server.Handler == nil {
		t.Fatalf("expected handler to be set")
	}
	if app.Core() == nil {
		t.Fatalf("expected core to be initialized")
	}
}

func TestNewReturnsErrorWhenDBParentDirMissing(t *testing.T) {
	t.Parallel()

	missingDBPath := filepath.Join(t.TempDir(), "missing", "vm-system.db")
	_, err := New(DefaultConfig(missingDBPath), http.NewServeMux())
	if err == nil {
		t.Fatalf("expected error for missing DB parent directory")
	}
	if !strings.Contains(err.Error(), "open store") {
		t.Fatalf("expected open store failure context, got %v", err)
	}
}
