package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

func newTemplateAddLibraryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-library <template-id>",
		Short: "Add a library to a template",
		Long:  "Add a library to a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.AddTemplateLibrary(context.Background(), templateID, vmclient.AddTemplateLibraryRequest{Name: name}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added library: %s to template %s\n", name, templateID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Library name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateRemoveLibraryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-library <template-id>",
		Short: "Remove a library from a template",
		Long:  "Remove a library from a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			if _, err := client.RemoveTemplateLibrary(context.Background(), templateID, name); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed library: %s from template %s\n", name, templateID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Library name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTemplateListLibrariesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-libraries <template-id>",
		Short: "List libraries configured on a template",
		Long:  "List libraries configured on a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			libraries, err := client.ListTemplateLibraries(context.Background(), templateID)
			if err != nil {
				return err
			}
			if len(libraries) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No libraries configured")
				return nil
			}
			for _, library := range libraries {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), library)
			}
			return nil
		},
	}
}
