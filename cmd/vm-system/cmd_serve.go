package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmdaemon"
)

func newServeCommand() *cobra.Command {
	var listenAddr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the vm-system daemon host",
		Long:  "Start a long-lived daemon process that hosts runtime sessions and serves API requests.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := vmdaemon.DefaultConfig(dbPath)
			if listenAddr != "" {
				cfg.ListenAddr = listenAddr
			}

			// Start with a minimal operational surface; API routes are added via transport package wiring.
			mux := http.NewServeMux()
			mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status": "ok",
				})
			})

			app, err := vmdaemon.New(cfg, mux)
			if err != nil {
				return err
			}
			defer app.Close()

			fmt.Printf("vm-system daemon listening on %s\n", cfg.ListenAddr)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			return app.Run(ctx)
		},
	}

	cmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:3210", "HTTP listen address")
	return cmd
}
