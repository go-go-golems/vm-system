package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

func newTemplateAddModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-module <template-id>",
		Short: "Add a module to a template",
		Long:  "Add a module to a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.AddTemplateModule(context.Background(), templateID, vmclient.AddTemplateModuleRequest{Name: name}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added module: %s to template %s\n", name, templateID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Module name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateRemoveModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-module <template-id>",
		Short: "Remove a module from a template",
		Long:  "Remove a module from a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.RemoveTemplateModule(context.Background(), templateID, name); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed module: %s from template %s\n", name, templateID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Module name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateListModulesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-modules <template-id>",
		Short: "List modules configured on a template",
		Long:  "List modules configured on a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			modules, err := client.ListTemplateModules(context.Background(), templateID)
			if err != nil {
				return err
			}
			if len(modules) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No modules configured")
				return nil
			}
			for _, module := range modules {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), module)
			}
			return nil
		},
	}
}
