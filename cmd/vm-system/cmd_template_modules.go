package main

import (
	"context"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

const (
	templateModulesActionAdd    = "add-module"
	templateModulesActionRemove = "remove-module"
	templateModulesActionList   = "list-modules"
)

type templateModulesCommand struct {
	*cmds.CommandDescription
	action string
}

var _ cmds.WriterCommand = &templateModulesCommand{}

func (c *templateModulesCommand) RunIntoWriter(_ context.Context, vals *values.Values, w io.Writer) error {
	client := vmclient.New(serverURL, nil)

	switch c.action {
	case templateModulesActionAdd:
		settings := &templateNameFlag{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		if _, err := client.AddTemplateModule(context.Background(), settings.TemplateID, vmclient.AddTemplateModuleRequest{Name: settings.Name}); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Added module: %s to template %s\n", settings.Name, settings.TemplateID)
		return nil
	case templateModulesActionRemove:
		settings := &templateNameFlag{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		if _, err := client.RemoveTemplateModule(context.Background(), settings.TemplateID, settings.Name); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Removed module: %s from template %s\n", settings.Name, settings.TemplateID)
		return nil
	case templateModulesActionList:
		settings := &templateIDArg{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		modules, err := client.ListTemplateModules(context.Background(), settings.TemplateID)
		if err != nil {
			return err
		}
		if len(modules) == 0 {
			_, _ = fmt.Fprintln(w, "No modules configured")
			return nil
		}
		for _, module := range modules {
			_, _ = fmt.Fprintln(w, module)
		}
		return nil
	default:
		return fmt.Errorf("unknown template modules action: %s", c.action)
	}
}

func newTemplateAddModuleCommand() *cobra.Command {
	command := &templateModulesCommand{
		CommandDescription: commandDescription(
			"add-module",
			"Add a module to a template",
			"Add a module to a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Module name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateModulesActionAdd,
	}
	return buildCobraCommand(command)
}

func newTemplateRemoveModuleCommand() *cobra.Command {
	command := &templateModulesCommand{
		CommandDescription: commandDescription(
			"remove-module",
			"Remove a module from a template",
			"Remove a module from a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Module name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateModulesActionRemove,
	}
	return buildCobraCommand(command)
}

func newTemplateListModulesCommand() *cobra.Command {
	command := &templateModulesCommand{
		CommandDescription: commandDescription(
			"list-modules",
			"List modules configured on a template",
			"List modules configured on a template.",
			nil,
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateModulesActionList,
	}
	return buildCobraCommand(command)
}
