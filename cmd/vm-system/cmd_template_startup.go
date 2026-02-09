package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

const (
	templateStartupActionAdd  = "add-startup"
	templateStartupActionList = "list-startup"
)

type templateStartupCommand struct {
	*cmds.CommandDescription
	action string
}

var _ cmds.WriterCommand = &templateStartupCommand{}

func (c *templateStartupCommand) RunIntoWriter(_ context.Context, vals *values.Values, w io.Writer) error {
	client := vmclient.New(serverURL, nil)

	switch c.action {
	case templateStartupActionAdd:
		settings := &templateAddStartupSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}
		mode := strings.ToLower(strings.TrimSpace(settings.Mode))
		if mode == "" {
			mode = "eval"
		}
		if mode != "eval" {
			return fmt.Errorf("unsupported startup mode %q: only eval is currently supported", settings.Mode)
		}
		settings.Mode = mode

		startup, err := client.AddTemplateStartupFile(context.Background(), settings.TemplateID, vmclient.AddTemplateStartupFileRequest{
			Path:       settings.Path,
			OrderIndex: settings.OrderIndex,
			Mode:       settings.Mode,
		})
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Added startup file: %s (order: %d) to template %s\n", startup.Path, startup.OrderIndex, settings.TemplateID)
		return nil
	case templateStartupActionList:
		settings := &templateIDArg{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		files, err := client.ListTemplateStartupFiles(context.Background(), settings.TemplateID)
		if err != nil {
			return err
		}

		if len(files) == 0 {
			_, _ = fmt.Fprintln(w, "No startup files found")
			return nil
		}

		_, _ = fmt.Fprintf(w, "%-5s %-10s %-50s\n", "Order", "Mode", "Path")
		_, _ = fmt.Fprintln(w, "----------------------------------------------------------------------")
		for _, file := range files {
			_, _ = fmt.Fprintf(w, "%-5d %-10s %-50s\n", file.OrderIndex, file.Mode, file.Path)
		}
		return nil
	default:
		return fmt.Errorf("unknown template startup action: %s", c.action)
	}
}

func newTemplateAddStartupFileCommand() *cobra.Command {
	command := &templateStartupCommand{
		CommandDescription: commandDescription(
			"add-startup",
			"Add a startup file to a template",
			"Add a startup file to a template.",
			[]*fields.Definition{
				fields.New("path", fields.TypeString, fields.WithRequired(true), fields.WithHelp("File path (required)")),
				fields.New("mode", fields.TypeString, fields.WithDefault("eval"), fields.WithHelp("Startup mode (only eval is currently supported)")),
				fields.New("order", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Order index")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateStartupActionAdd,
	}
	return buildCobraCommand(command)
}

func newTemplateListStartupFilesCommand() *cobra.Command {
	command := &templateStartupCommand{
		CommandDescription: commandDescription(
			"list-startup",
			"List startup files for a template",
			"List startup files for a template.",
			nil,
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateStartupActionList,
	}
	return buildCobraCommand(command)
}
