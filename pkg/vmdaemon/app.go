package vmdaemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-go-golems/vm-system/pkg/vmcontrol"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

// App hosts the long-lived runtime process around vmcontrol and HTTP transport.
type App struct {
	cfg    Config
	store  *vmstore.VMStore
	core   *vmcontrol.Core
	server *http.Server
}

func New(cfg Config, handler http.Handler) (*App, error) {
	store, err := vmstore.NewVMStore(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
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
