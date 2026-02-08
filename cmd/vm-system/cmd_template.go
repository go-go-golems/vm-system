package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

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
	var name, engine string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new template",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			template, err := client.CreateTemplate(context.Background(), vmclient.CreateTemplateRequest{
				Name:   name,
				Engine: engine,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Created template: %s (ID: %s)\n", template.Name, template.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Template name (required)")
	cmd.Flags().StringVar(&engine, "engine", "goja", "Engine type (goja, quickjs, node, custom)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templates, err := client.ListTemplates(context.Background())
			if err != nil {
				return err
			}

			if len(templates) == 0 {
				fmt.Println("No templates found")
				return nil
			}

			fmt.Printf("%-36s %-20s %-10s %-10s\n", "ID", "Name", "Engine", "Active")
			fmt.Println("------------------------------------------------------------------------------------")
			for _, template := range templates {
				active := "no"
				if template.IsActive {
					active = "yes"
				}
				fmt.Printf("%-36s %-20s %-10s %-10s\n", template.ID, template.Name, template.Engine, active)
			}

			return nil
		},
	}
}

func newTemplateGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [template-id]",
		Short: "Get template details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			detail, err := client.GetTemplate(context.Background(), templateID)
			if err != nil {
				return err
			}

			template := detail.Template
			fmt.Printf("Template: %s\n", template.Name)
			fmt.Printf("ID: %s\n", template.ID)
			fmt.Printf("Engine: %s\n", template.Engine)
			fmt.Printf("Active: %v\n", template.IsActive)
			fmt.Printf("Created: %s\n", template.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Updated: %s\n", template.UpdatedAt.Format(time.RFC3339))

			if detail.Settings != nil {
				fmt.Println("\nSettings:")
				fmt.Printf("  Limits: %s\n", string(detail.Settings.Limits))
				fmt.Printf("  Resolver: %s\n", string(detail.Settings.Resolver))
				fmt.Printf("  Runtime: %s\n", string(detail.Settings.Runtime))
			}

			if len(detail.Capabilities) > 0 {
				fmt.Println("\nCapabilities:")
				for _, capability := range detail.Capabilities {
					enabled := "disabled"
					if capability.Enabled {
						enabled = "enabled"
					}
					fmt.Printf("  [%s] %s:%s (config: %s)\n", enabled, capability.Kind, capability.Name, string(capability.Config))
				}
			}

			if len(detail.StartupFiles) > 0 {
				fmt.Println("\nStartup Files:")
				for _, startup := range detail.StartupFiles {
					fmt.Printf("  [%d] %s (mode: %s)\n", startup.OrderIndex, startup.Path, startup.Mode)
				}
			}

			if len(template.ExposedModules) > 0 {
				fmt.Println("\nExposed Modules:")
				for _, module := range template.ExposedModules {
					fmt.Printf("  - %s\n", module)
				}
			}

			if len(template.Libraries) > 0 {
				fmt.Println("\nLoaded Libraries:")
				for _, library := range template.Libraries {
					fmt.Printf("  - %s\n", library)
				}
			}

			return nil
		},
	}
}

func newTemplateDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [template-id]",
		Short: "Delete a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]
			if err := client.DeleteTemplate(context.Background(), templateID); err != nil {
				return err
			}
			fmt.Printf("Deleted template: %s\n", templateID)
			return nil
		},
	}
}

func newTemplateAddModuleCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add-module [template-id]",
		Short: "Add a module to a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			if _, err := client.AddTemplateModule(context.Background(), templateID, vmclient.AddTemplateModuleRequest{
				Name: name,
			}); err != nil {
				return err
			}

			fmt.Printf("Added module: %s to template %s\n", name, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Module name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateRemoveModuleCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "remove-module [template-id]",
		Short: "Remove a module from a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			if _, err := client.RemoveTemplateModule(context.Background(), templateID, name); err != nil {
				return err
			}

			fmt.Printf("Removed module: %s from template %s\n", name, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Module name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateListModulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-modules [template-id]",
		Short: "List modules configured on a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			modules, err := client.ListTemplateModules(context.Background(), templateID)
			if err != nil {
				return err
			}

			if len(modules) == 0 {
				fmt.Println("No modules configured")
				return nil
			}

			for _, module := range modules {
				fmt.Println(module)
			}

			return nil
		},
	}
}

func newTemplateAddLibraryCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add-library [template-id]",
		Short: "Add a library to a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			if _, err := client.AddTemplateLibrary(context.Background(), templateID, vmclient.AddTemplateLibraryRequest{
				Name: name,
			}); err != nil {
				return err
			}

			fmt.Printf("Added library: %s to template %s\n", name, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Library name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateRemoveLibraryCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "remove-library [template-id]",
		Short: "Remove a library from a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			if _, err := client.RemoveTemplateLibrary(context.Background(), templateID, name); err != nil {
				return err
			}

			fmt.Printf("Removed library: %s from template %s\n", name, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Library name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateListLibrariesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-libraries [template-id]",
		Short: "List libraries configured on a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			libraries, err := client.ListTemplateLibraries(context.Background(), templateID)
			if err != nil {
				return err
			}

			if len(libraries) == 0 {
				fmt.Println("No libraries configured")
				return nil
			}

			for _, library := range libraries {
				fmt.Println(library)
			}

			return nil
		},
	}
}

func newTemplateListAvailableModulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-available-modules",
		Short: "List built-in module catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			modules := vmmodels.BuiltinModules()
			data, err := json.MarshalIndent(modules, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newTemplateListAvailableLibrariesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-available-libraries",
		Short: "List built-in library catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			libraries := vmmodels.BuiltinLibraries()
			data, err := json.MarshalIndent(libraries, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newTemplateAddCapabilityCommand() *cobra.Command {
	var kind, name, configJSON string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "add-capability [template-id]",
		Short: "Add a capability to a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			var config map[string]interface{}
			if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
				return fmt.Errorf("invalid config JSON: %w", err)
			}

			capability, err := client.AddTemplateCapability(context.Background(), templateID, vmclient.AddTemplateCapabilityRequest{
				Kind:    kind,
				Name:    name,
				Enabled: enabled,
				Config:  config,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Added capability: %s:%s to template %s\n", capability.Kind, capability.Name, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "module", "Capability kind (module, global, fs, net, env)")
	cmd.Flags().StringVar(&name, "name", "", "Capability name (required)")
	cmd.Flags().StringVar(&configJSON, "config", "{}", "Capability config (JSON)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable capability")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newTemplateListCapabilitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-capabilities [template-id]",
		Short: "List capabilities for a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			capabilities, err := client.ListTemplateCapabilities(context.Background(), templateID)
			if err != nil {
				return err
			}

			if len(capabilities) == 0 {
				fmt.Println("No capabilities found")
				return nil
			}

			fmt.Printf("%-10s %-20s %-10s %-30s\n", "Kind", "Name", "Enabled", "Config")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, capability := range capabilities {
				enabledText := "no"
				if capability.Enabled {
					enabledText = "yes"
				}
				configText := string(capability.Config)
				if len(configText) > 30 {
					configText = configText[:27] + "..."
				}
				fmt.Printf("%-10s %-20s %-10s %-30s\n", capability.Kind, capability.Name, enabledText, configText)
			}

			return nil
		},
	}
}

func newTemplateAddStartupFileCommand() *cobra.Command {
	var path, mode string
	var orderIndex int

	cmd := &cobra.Command{
		Use:   "add-startup [template-id]",
		Short: "Add a startup file to a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			startup, err := client.AddTemplateStartupFile(context.Background(), templateID, vmclient.AddTemplateStartupFileRequest{
				Path:       path,
				OrderIndex: orderIndex,
				Mode:       mode,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Added startup file: %s (order: %d) to template %s\n", startup.Path, startup.OrderIndex, templateID)
			return nil
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "File path (required)")
	cmd.Flags().StringVar(&mode, "mode", "eval", "Mode (eval or import)")
	cmd.Flags().IntVar(&orderIndex, "order", 10, "Order index")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func newTemplateListStartupFilesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-startup [template-id]",
		Short: "List startup files for a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			templateID := args[0]

			files, err := client.ListTemplateStartupFiles(context.Background(), templateID)
			if err != nil {
				return err
			}

			if len(files) == 0 {
				fmt.Println("No startup files found")
				return nil
			}

			fmt.Printf("%-5s %-10s %-50s\n", "Order", "Mode", "Path")
			fmt.Println("----------------------------------------------------------------------")
			for _, file := range files {
				fmt.Printf("%-5d %-10s %-50s\n", file.OrderIndex, file.Mode, file.Path)
			}

			return nil
		},
	}
}
