package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmdaemon"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

type serveSettings struct {
	ListenAddr string `glazed:"listen"`
}

func newServeCommand() *cobra.Command {
	cmd := &bareCommand{
		CommandDescription: mustCommandDescription(
			"serve",
			"Run the vm-system daemon host",
			"Start a long-lived daemon process that hosts runtime sessions and serves API requests.",
			[]*fields.Definition{
				fields.New("listen", fields.TypeString, fields.WithDefault("127.0.0.1:3210"), fields.WithHelp("HTTP listen address")),
			},
			nil,
			false,
		),
		run: func(_ context.Context, vals *values.Values) error {
			settings := &serveSettings{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			cfg := vmdaemon.DefaultConfig(dbPath)
			cfg.ListenAddr = settings.ListenAddr

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

	return mustBuildCobraCommand(cmd)
}
