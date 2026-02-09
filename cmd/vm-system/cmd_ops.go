package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
)

func newOpsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ops",
		Short: "Operational daemon commands",
		Long:  "Read daemon health and runtime summary endpoints.",
	}

	cmd.AddCommand(
		newOpsHealthCommand(),
		newOpsRuntimeSummaryCommand(),
	)

	return cmd
}

func newOpsHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Get daemon health",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			status, err := client.Health(context.Background())
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}

func newOpsRuntimeSummaryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "runtime-summary",
		Short: "Get daemon runtime summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			summary, err := client.RuntimeSummary(context.Background())
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		},
	}
}
