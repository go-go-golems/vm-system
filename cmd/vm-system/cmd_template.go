package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/spf13/cobra"
)

type templateCreateSettings struct {
	Name   string `glazed:"name"`
	Engine string `glazed:"engine"`
}

type templateIDArg struct {
	TemplateID string `glazed:"template-id"`
}

type templateNameFlag struct {
	TemplateID string `glazed:"template-id"`
	Name       string `glazed:"name"`
}

type templateAddCapabilitySettings struct {
	TemplateID string `glazed:"template-id"`
	Kind       string `glazed:"kind"`
	Name       string `glazed:"name"`
	ConfigJSON string `glazed:"config"`
	Enabled    bool   `glazed:"enabled"`
}

type templateAddStartupSettings struct {
	TemplateID string `glazed:"template-id"`
	Path       string `glazed:"path"`
	Mode       string `glazed:"mode"`
	OrderIndex int    `glazed:"order"`
}

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates via daemon API",
		Long:  `Create, list, inspect, and delete templates plus capability/startup policy metadata through the daemon REST API.`,
	}

	cmd.AddCommand(
		newTemplateCreateCommand(),
		newTemplateListCommand(),
		newTemplateGetCommand(),
		newTemplateDeleteCommand(),
		newTemplateAddModuleCommand(),
		newTemplateRemoveModuleCommand(),
		newTemplateListModulesCommand(),
		newTemplateAddLibraryCommand(),
		newTemplateRemoveLibraryCommand(),
		newTemplateListLibrariesCommand(),
		newTemplateListAvailableModulesCommand(),
		newTemplateListAvailableLibrariesCommand(),
		newTemplateAddCapabilityCommand(),
		newTemplateListCapabilitiesCommand(),
		newTemplateAddStartupFileCommand(),
		newTemplateListStartupFilesCommand(),
	)

	return cmd
}

func newTemplateCreateCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"create",
			"Create a new template",
			"Create a new runtime template.",
			[]*fields.Definition{
				fields.New("name", fields.TypeString, fields.WithHelp("Template name (required)"), fields.WithRequired(true)),
				fields.New("engine", fields.TypeString, fields.WithDefault("goja"), fields.WithHelp("Engine type (goja, quickjs, node, custom)")),
			},
			nil,
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateCreateSettings{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			template, err := client.CreateTemplate(context.Background(), vmclient.CreateTemplateRequest{
				Name:   settings.Name,
				Engine: settings.Engine,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Created template: %s (ID: %s)\n", template.Name, template.ID)
			return nil
		},
	}

	return mustBuildCobraCommand(command)
}

func newTemplateListCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list",
			"List templates",
			"List all templates.",
			nil,
			nil,
			false,
		),
		run: func(_ context.Context, _ *values.Values, w io.Writer) error {
			client := vmclient.New(serverURL, nil)
			templates, err := client.ListTemplates(context.Background())
			if err != nil {
				return err
			}

			if len(templates) == 0 {
				_, _ = fmt.Fprintln(w, "No templates found")
				return nil
			}

			_, _ = fmt.Fprintf(w, "%-36s %-20s %-10s %-10s\n", "ID", "Name", "Engine", "Active")
			_, _ = fmt.Fprintln(w, "------------------------------------------------------------------------------------")
			for _, template := range templates {
				active := "no"
				if template.IsActive {
					active = "yes"
				}
				_, _ = fmt.Fprintf(w, "%-36s %-20s %-10s %-10s\n", template.ID, template.Name, template.Engine, active)
			}

			return nil
		},
	}

	return mustBuildCobraCommand(command)
}

func newTemplateGetCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"get",
			"Get template details",
			"Get template details by template ID.",
			nil,
			[]*fields.Definition{
				fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID")),
			},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateIDArg{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			detail, err := client.GetTemplate(context.Background(), settings.TemplateID)
			if err != nil {
				return err
			}

			template := detail.Template
			_, _ = fmt.Fprintf(w, "Template: %s\n", template.Name)
			_, _ = fmt.Fprintf(w, "ID: %s\n", template.ID)
			_, _ = fmt.Fprintf(w, "Engine: %s\n", template.Engine)
			_, _ = fmt.Fprintf(w, "Active: %v\n", template.IsActive)
			_, _ = fmt.Fprintf(w, "Created: %s\n", template.CreatedAt.Format(time.RFC3339))
			_, _ = fmt.Fprintf(w, "Updated: %s\n", template.UpdatedAt.Format(time.RFC3339))

			if detail.Settings != nil {
				_, _ = fmt.Fprintln(w, "\nSettings:")
				_, _ = fmt.Fprintf(w, "  Limits: %s\n", string(detail.Settings.Limits))
				_, _ = fmt.Fprintf(w, "  Resolver: %s\n", string(detail.Settings.Resolver))
				_, _ = fmt.Fprintf(w, "  Runtime: %s\n", string(detail.Settings.Runtime))
			}

			if len(detail.Capabilities) > 0 {
				_, _ = fmt.Fprintln(w, "\nCapabilities:")
				for _, capability := range detail.Capabilities {
					enabled := "disabled"
					if capability.Enabled {
						enabled = "enabled"
					}
					_, _ = fmt.Fprintf(w, "  [%s] %s:%s (config: %s)\n", enabled, capability.Kind, capability.Name, string(capability.Config))
				}
			}

			if len(detail.StartupFiles) > 0 {
				_, _ = fmt.Fprintln(w, "\nStartup Files:")
				for _, startup := range detail.StartupFiles {
					_, _ = fmt.Fprintf(w, "  [%d] %s (mode: %s)\n", startup.OrderIndex, startup.Path, startup.Mode)
				}
			}

			if len(template.ExposedModules) > 0 {
				_, _ = fmt.Fprintln(w, "\nExposed Modules:")
				for _, module := range template.ExposedModules {
					_, _ = fmt.Fprintf(w, "  - %s\n", module)
				}
			}

			if len(template.Libraries) > 0 {
				_, _ = fmt.Fprintln(w, "\nLoaded Libraries:")
				for _, library := range template.Libraries {
					_, _ = fmt.Fprintf(w, "  - %s\n", library)
				}
			}

			return nil
		},
	}

	return mustBuildCobraCommand(command)
}

func newTemplateDeleteCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"delete",
			"Delete a template",
			"Delete a template by template ID.",
			nil,
			[]*fields.Definition{
				fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID")),
			},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateIDArg{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if err := client.DeleteTemplate(context.Background(), settings.TemplateID); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Deleted template: %s\n", settings.TemplateID)
			return nil
		},
	}

	return mustBuildCobraCommand(command)
}

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

func newTemplateListAvailableModulesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-available-modules",
			"List built-in module catalog",
			"List built-in module catalog.",
			nil,
			nil,
			false,
		),
		run: func(_ context.Context, _ *values.Values, w io.Writer) error {
			modules := vmmodels.BuiltinModules()
			data, err := json.MarshalIndent(modules, "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(w, string(data))
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateListAvailableLibrariesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-available-libraries",
			"List built-in library catalog",
			"List built-in library catalog.",
			nil,
			nil,
			false,
		),
		run: func(_ context.Context, _ *values.Values, w io.Writer) error {
			libraries := vmmodels.BuiltinLibraries()
			data, err := json.MarshalIndent(libraries, "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(w, string(data))
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateAddCapabilityCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"add-capability",
			"Add a capability to a template",
			"Add a capability to a template.",
			[]*fields.Definition{
				fields.New("kind", fields.TypeString, fields.WithDefault("module"), fields.WithHelp("Capability kind (module, global, fs, net, env)")),
				fields.New("name", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Capability name (required)")),
				fields.New("config", fields.TypeString, fields.WithDefault("{}"), fields.WithHelp("Capability config (JSON)")),
				fields.New("enabled", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Enable capability")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateAddCapabilitySettings{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			var config map[string]interface{}
			if err := json.Unmarshal([]byte(settings.ConfigJSON), &config); err != nil {
				return fmt.Errorf("invalid config JSON: %w", err)
			}

			client := vmclient.New(serverURL, nil)
			capability, err := client.AddTemplateCapability(context.Background(), settings.TemplateID, vmclient.AddTemplateCapabilityRequest{
				Kind:    settings.Kind,
				Name:    settings.Name,
				Enabled: settings.Enabled,
				Config:  config,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Added capability: %s:%s to template %s\n", capability.Kind, capability.Name, settings.TemplateID)
			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateListCapabilitiesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-capabilities",
			"List capabilities for a template",
			"List capabilities for a template.",
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
			capabilities, err := client.ListTemplateCapabilities(context.Background(), settings.TemplateID)
			if err != nil {
				return err
			}

			if len(capabilities) == 0 {
				_, _ = fmt.Fprintln(w, "No capabilities found")
				return nil
			}

			_, _ = fmt.Fprintf(w, "%-10s %-20s %-10s %-30s\n", "Kind", "Name", "Enabled", "Config")
			_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
			for _, capability := range capabilities {
				enabledText := "no"
				if capability.Enabled {
					enabledText = "yes"
				}
				configText := string(capability.Config)
				if len(configText) > 30 {
					configText = configText[:27] + "..."
				}
				_, _ = fmt.Fprintf(w, "%-10s %-20s %-10s %-30s\n", capability.Kind, capability.Name, enabledText, configText)
			}

			return nil
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateAddStartupFileCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"add-startup",
			"Add a startup file to a template",
			"Add a startup file to a template.",
			[]*fields.Definition{
				fields.New("path", fields.TypeString, fields.WithRequired(true), fields.WithHelp("File path (required)")),
				fields.New("mode", fields.TypeString, fields.WithDefault("eval"), fields.WithHelp("Mode (eval or import)")),
				fields.New("order", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Order index")),
			},
			[]*fields.Definition{fields.New("template-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Template ID"))},
			false,
		),
		run: func(_ context.Context, vals *values.Values, w io.Writer) error {
			settings := &templateAddStartupSettings{}
			if err := decodeDefault(vals, settings); err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
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
		},
	}
	return mustBuildCobraCommand(command)
}

func newTemplateListStartupFilesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-startup",
			"List startup files for a template",
			"List startup files for a template.",
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
		},
	}
	return mustBuildCobraCommand(command)
}
