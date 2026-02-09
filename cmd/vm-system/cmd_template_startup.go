package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

func newTemplateAddStartupFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-startup <template-id>",
		Short: "Add a startup file to a template",
		Long:  "Add a startup file to a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			path, err := cmd.Flags().GetString("path")
			if err != nil {
				return err
			}
			mode, err := cmd.Flags().GetString("mode")
			if err != nil {
				return err
			}
			orderIndex, err := cmd.Flags().GetInt("order")
			if err != nil {
				return err
			}

			mode = strings.ToLower(strings.TrimSpace(mode))
			if mode == "" {
				mode = "eval"
			}
			if mode != "eval" {
				return fmt.Errorf("unsupported startup mode %q: only eval is currently supported", mode)
			}

			client := vmclient.New(serverURL, nil)
			startup, err := client.AddTemplateStartupFile(context.Background(), templateID, vmclient.AddTemplateStartupFileRequest{
				Path:       path,
				OrderIndex: orderIndex,
				Mode:       mode,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added startup file: %s (order: %d) to template %s\n", startup.Path, startup.OrderIndex, templateID)
			return nil
		},
	}
	cmd.Flags().String("path", "", "File path (required)")
	cmd.Flags().String("mode", "eval", "Startup mode (only eval is currently supported)")
	cmd.Flags().Int("order", 10, "Order index")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func newTemplateListStartupFilesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-startup <template-id>",
		Short: "List startup files for a template",
		Long:  "List startup files for a template.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateID := args[0]

			client := vmclient.New(serverURL, nil)
			files, err := client.ListTemplateStartupFiles(context.Background(), templateID)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
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
}
