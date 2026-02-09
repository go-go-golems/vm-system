package vmdaemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
	"github.com/rs/zerolog/log"
)

// App hosts the long-lived runtime process around vmcontrol and HTTP transport.
type App struct {
	cfg    Config
	store  *vmstore.VMStore
	core   *vmcontrol.Core
	server *http.Server
}

const sessionStartupGCMessage = "garbage collected on daemon startup: runtime state does not survive process restarts"

func New(cfg Config, handler http.Handler) (*App, error) {
	store, err := vmstore.NewVMStore(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	if err := closeStaleSessionsOnStartup(store); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("reconcile stale sessions on startup: %w", err)
	}

	core := vmcontrol.NewCore(store)
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTime,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		store:  store,
		core:   core,
		server: server,
	}, nil
}

func closeStaleSessionsOnStartup(store *vmstore.VMStore) error {
	sessions, err := store.ListSessions("")
	if err != nil {
		return err
	}

	closedCount := 0
	now := time.Now()

	for _, session := range sessions {
		if session.Status != string(vmmodels.SessionStarting) && session.Status != string(vmmodels.SessionReady) {
			continue
		}

		session.Status = string(vmmodels.SessionClosed)
		if session.ClosedAt == nil {
			closedAt := now
			session.ClosedAt = &closedAt
		}
		if session.LastError == "" {
			session.LastError = sessionStartupGCMessage
		}

		if err := store.UpdateSession(session); err != nil {
			return fmt.Errorf("update stale session %s: %w", session.ID, err)
		}

		closedCount++
	}

	if closedCount > 0 {
		log.Warn().
			Int("closed_sessions", closedCount).
			Msg("closed stale persisted sessions on daemon startup")
	}

	return nil
}

func (a *App) Core() *vmcontrol.Core {
	return a.core
}

func (a *App) SetHandler(handler http.Handler) {
	a.server.Handler = handler
}

// Run starts the server and blocks until context cancellation or server error.
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	}
}

func (a *App) Close() error {
	return a.store.Close()
}
