package main

import (
	"context"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

func newTemplateAddModuleCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"add-module",
			"Add a module to a template",
			"Add a module to a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Module name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateNameFlag{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.AddTemplateModule(context.Background(), settings.TemplateID, vmclient.AddTemplateModuleRequest{Name: settings.Name}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Added module: %s to template %s\n", settings.Name, settings.TemplateID)
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateRemoveModuleCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"remove-module",
			"Remove a module from a template",
			"Remove a module from a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Module name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateNameFlag{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.RemoveTemplateModule(context.Background(), settings.TemplateID, settings.Name); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Removed module: %s from template %s\n", settings.Name, settings.TemplateID)
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateListModulesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-modules",
			"List modules configured on a template",
			"List modules configured on a template.",
			nil,
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateIDArg{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
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
		},
	}
	return mustBuildCobraCommand(command)
}
