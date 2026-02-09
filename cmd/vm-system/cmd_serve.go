package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmdaemon"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the vm-system daemon host",
		Long:  "Start a long-lived daemon process that hosts runtime sessions and serves API requests.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			listenAddr, err := cmd.Flags().GetString("listen")
			if err != nil {
				return err
			}

			cfg := vmdaemon.DefaultConfig(dbPath)
			cfg.ListenAddr = listenAddr

			app, err := vmdaemon.New(cfg, http.NewServeMux())
			if err != nil {
				return err
			}
			defer app.Close()

			app.SetHandler(vmhttp.NewHandler(app.Core()))
			fmt.Printf("vm-system daemon listening on %s\n", cfg.ListenAddr)

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			return app.Run(ctx)
		},
	}

	cmd.Flags().String("listen", "127.0.0.1:3210", "HTTP listen address")

	return cmd
}
