package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/internal/web"
	"github.com/go-go-golems/vm-system/pkg/vmdaemon"
	vmhttp "github.com/go-go-golems/vm-system/pkg/vmtransport/http"
)

type serveSettings struct {
	ListenAddr string `glazed:"listen"`
}

type serveCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &serveCommand{}

func (c *serveCommand) Run(_ context.Context, vals *values.Values) error {
	settings := &serveSettings{}
	if err := decodeDefault(vals, settings); err != nil {
		return err
	}

	cfg := vmdaemon.DefaultConfig(dbPath)
	cfg.ListenAddr = settings.ListenAddr

	app, err := vmdaemon.New(cfg, nil)
	if err != nil {
		return err
	}
	defer app.Close()

	apiHandler := vmhttp.NewHandler(app.Core())
	publicFS, fsErr := web.PublicFS()
	if fsErr != nil {
		app.SetHandler(apiHandler)
		log.Warn().
			Err(fsErr).
			Msg("web ui assets unavailable; serving API only")
	} else {
		app.SetHandler(web.NewHandler(apiHandler, publicFS))
		log.Info().
			Str("component", "daemon").
			Msg("web ui assets enabled")
	}

	log.Info().
		Str("component", "daemon").
		Str("listen_addr", cfg.ListenAddr).
		Msg("vm-system daemon listening")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return app.Run(ctx)
}

func newServeCommand() *cobra.Command {
	return buildCobraCommand(&serveCommand{
		CommandDescription: commandDescription(
			"serve",
			"Run the vm-system daemon host",
			"Start a long-lived daemon process that hosts runtime sessions and serves API requests.",
			[]*fields.Definition{
				fields.New("listen", fields.TypeString, fields.WithDefault("127.0.0.1:3210"), fields.WithHelp("HTTP listen address")),
			},
			nil,
			false,
		),
	})
}
