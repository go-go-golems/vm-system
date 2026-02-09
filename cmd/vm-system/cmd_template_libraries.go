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

func newTemplateAddLibraryCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"add-library",
			"Add a library to a template",
			"Add a library to a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Library name (required)")),
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
			if _, err := client.AddTemplateLibrary(context.Background(), settings.TemplateID, vmclient.AddTemplateLibraryRequest{Name: settings.Name}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Added library: %s to template %s\n", settings.Name, settings.TemplateID)
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateRemoveLibraryCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"remove-library",
			"Remove a library from a template",
			"Remove a library from a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Library name (required)")),
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
			if _, err := client.RemoveTemplateLibrary(context.Background(), settings.TemplateID, settings.Name); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Removed library: %s from template %s\n", settings.Name, settings.TemplateID)
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateListLibrariesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-libraries",
			"List libraries configured on a template",
			"List libraries configured on a template.",
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
			libraries, err := client.ListTemplateLibraries(context.Background(), settings.TemplateID)
			if err != nil {
				return err
			}
			if len(libraries) == 0 {
				_, _ = fmt.Fprintln(w, "No libraries configured")
				return nil
			}
			for _, library := range libraries {
				_, _ = fmt.Fprintln(w, library)
			}
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}
