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

func newTemplateListAvailableModulesCommand() *cobra.Command {
	command := &writerCommand{
		CommandDescription: mustCommandDescription(
			"list-available-modules",
			"List configurable native module catalog",
			"List configurable native module catalog (JavaScript built-ins are always available and not template-configurable).",
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
