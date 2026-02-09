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
	templateLibrariesActionAdd    = "add-library"
	templateLibrariesActionRemove = "remove-library"
	templateLibrariesActionList   = "list-libraries"
)

type templateLibrariesCommand struct {
	*cmds.CommandDescription
	action string
}

var _ cmds.WriterCommand = &templateLibrariesCommand{}

func (c *templateLibrariesCommand) RunIntoWriter(_ context.Context, vals *values.Values, w io.Writer) error {
	client := vmclient.New(serverURL, nil)

	switch c.action {
	case templateLibrariesActionAdd:
		settings := &templateNameFlag{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		if _, err := client.AddTemplateLibrary(context.Background(), settings.TemplateID, vmclient.AddTemplateLibraryRequest{Name: settings.Name}); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Added library: %s to template %s\n", settings.Name, settings.TemplateID)
		return nil
	case templateLibrariesActionRemove:
		settings := &templateNameFlag{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		if _, err := client.RemoveTemplateLibrary(context.Background(), settings.TemplateID, settings.Name); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Removed library: %s from template %s\n", settings.Name, settings.TemplateID)
		return nil
	case templateLibrariesActionList:
		settings := &templateIDArg{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

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
	default:
		return fmt.Errorf("unknown template libraries action: %s", c.action)
	}
}

func newTemplateAddLibraryCommand() *cobra.Command {
	command := &templateLibrariesCommand{
		CommandDescription: commandDescription(
			"add-library",
			"Add a library to a template",
			"Add a library to a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Library name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateLibrariesActionAdd,
	}
	return buildCobraCommand(command)
}

func newTemplateRemoveLibraryCommand() *cobra.Command {
	command := &templateLibrariesCommand{
		CommandDescription: commandDescription(
			"remove-library",
			"Remove a library from a template",
			"Remove a library from a template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Library name (required)")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateLibrariesActionRemove,
	}
	return buildCobraCommand(command)
}

func newTemplateListLibrariesCommand() *cobra.Command {
	command := &templateLibrariesCommand{
		CommandDescription: commandDescription(
			"list-libraries",
			"List libraries configured on a template",
			"List libraries configured on a template.",
			nil,
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		action: templateLibrariesActionList,
	}
	return buildCobraCommand(command)
}
