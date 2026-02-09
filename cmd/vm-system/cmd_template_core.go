package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/spf13/cobra"
)

func newTemplateCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new template",
		Long:  "Create a new runtime template.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			engine, err := cmd.Flags().GetString("engine")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			template, err := client.CreateTemplate(context.Background(), vmclient.CreateTemplateRequest{
				Name:   name,
				Engine: engine,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created template: %s (ID: %s)\n", template.Name, template.ID)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Template name (required)")
	cmd.Flags().String("engine", "goja", "Engine type (goja, quickjs, node, custom)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List templates",
		Long:  "List all templates.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := vmclient.New(serverURL, nil)
			templates, err := client.ListTemplates(context.Background())
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
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
}

func newTemplateGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <template-id>",
		Short: "Get template details",
		Long:  "Get template details by template ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			detail, err := client.GetTemplate(context.Background(), templateID)
			if err != nil {
				return err
			}

			template := detail.Template
			w := cmd.OutOrStdout()
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
}

func newTemplateDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <template-id>",
		Short: "Delete a template",
		Long:  "Delete a template by template ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			if err := client.DeleteTemplate(context.Background(), templateID); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted template: %s\n", templateID)
			return nil
		},
	}
}

func newTemplateListAvailableModulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-available-modules",
		Short: "List configurable native module catalog",
		Long:  "List configurable native module catalog (JavaScript built-ins are always available and not template-configurable).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			modules := vmmodels.BuiltinModules()
			data, err := json.MarshalIndent(modules, "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newTemplateListAvailableLibrariesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-available-libraries",
		Short: "List built-in library catalog",
		Long:  "List built-in library catalog.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			libraries := vmmodels.BuiltinLibraries()
			data, err := json.MarshalIndent(libraries, "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newTemplateAddCapabilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-capability <template-id>",
		Short: "Add a capability to a template",
		Long:  "Add a capability to a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			kind, err := cmd.Flags().GetString("kind")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			configJSON, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			enabled, err := cmd.Flags().GetBool("enabled")
			if err != nil {
				return err
			}

			var config map[string]interface{}
			if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
				return fmt.Errorf("invalid config JSON: %w", err)
			}

			client := vmclient.New(serverURL, nil)
			capability, err := client.AddTemplateCapability(context.Background(), templateID, vmclient.AddTemplateCapabilityRequest{
				Kind:    kind,
				Name:    name,
				Enabled: enabled,
				Config:  config,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added capability: %s:%s to template %s\n", capability.Kind, capability.Name, templateID)
			return nil
		},
	}

	cmd.Flags().String("kind", "module", "Capability kind (module, global, fs, net, env)")
	cmd.Flags().String("name", "", "Capability name (required)")
	cmd.Flags().String("config", "{}", "Capability config (JSON)")
	cmd.Flags().Bool("enabled", true, "Enable capability")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateListCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-capabilities <template-id>",
		Short: "List capabilities for a template",
		Long:  "List capabilities for a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			capabilities, err := client.ListTemplateCapabilities(context.Background(), templateID)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
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
}
