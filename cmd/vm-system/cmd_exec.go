package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
)

func newExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute code in VM sessions via daemon API",
		Long:  `Execute REPL snippets or run files in VM sessions through the daemon REST API.`,
	}

	cmd.AddCommand(
		newExecREPLCommand(),
		newExecRunFileCommand(),
		newExecListCommand(),
		newExecGetCommand(),
		newExecEventsCommand(),
	)

	return cmd
}

func newExecREPLCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "repl [session-id] [code]",
		Short: "Execute REPL code",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessionID := args[0]
			code := args[1]

			execution, err := client.ExecuteREPL(context.Background(), vmclient.ExecuteREPLRequest{
				SessionID: sessionID,
				Input:     code,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", execution.ID)
			fmt.Printf("Status: %s\n", execution.Status)

			if execution.Result != nil {
				fmt.Printf("Result: %s\n", string(execution.Result))
			}

			if execution.Error != nil {
				fmt.Printf("Error: %s\n", string(execution.Error))
			}

			events, err := client.GetExecutionEvents(context.Background(), execution.ID, 0)
			if err != nil {
				return err
			}

			if len(events) > 0 {
				fmt.Println("\nEvents:")
				for _, event := range events {
					fmt.Printf("[%d] %s: %s\n", event.Seq, event.Type, string(event.Payload))
				}
			}

			return nil
		},
	}
}

func newExecRunFileCommand() *cobra.Command {
	var argsJSON, envJSON string

	cmd := &cobra.Command{
		Use:   "run-file [session-id] [path]",
		Short: "Run a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessionID := args[0]
			path := args[1]

			var argsMap map[string]interface{}
			var envMap map[string]interface{}

			if argsJSON != "" {
				if err := json.Unmarshal([]byte(argsJSON), &argsMap); err != nil {
					return fmt.Errorf("invalid args JSON: %w", err)
				}
			}

			if envJSON != "" {
				if err := json.Unmarshal([]byte(envJSON), &envMap); err != nil {
					return fmt.Errorf("invalid env JSON: %w", err)
				}
			}

			execution, err := client.ExecuteRunFile(context.Background(), vmclient.ExecuteRunFileRequest{
				SessionID: sessionID,
				Path:      path,
				Args:      argsMap,
				Env:       envMap,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", execution.ID)
			fmt.Printf("Status: %s\n", execution.Status)

			if execution.Error != nil {
				fmt.Printf("Error: %s\n", string(execution.Error))
			}

			events, err := client.GetExecutionEvents(context.Background(), execution.ID, 0)
			if err != nil {
				return err
			}

			if len(events) > 0 {
				fmt.Println("\nEvents:")
				for _, event := range events {
					fmt.Printf("[%d] %s: %s\n", event.Seq, event.Type, string(event.Payload))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&argsJSON, "args", "{}", "Arguments as JSON")
	cmd.Flags().StringVar(&envJSON, "env", "{}", "Environment as JSON")
	return cmd
}

func newExecListCommand() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list [session-id]",
		Short: "List executions for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessionID := args[0]

			executions, err := client.ListExecutions(context.Background(), sessionID, limit)
			if err != nil {
				return err
			}

			if len(executions) == 0 {
				fmt.Println("No executions found")
				return nil
			}

			fmt.Printf("%-36s %-10s %-10s %-20s\n", "Execution ID", "Kind", "Status", "Started")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, execution := range executions {
				fmt.Printf("%-36s %-10s %-10s %-20s\n",
					execution.ID,
					execution.Kind,
					execution.Status,
					execution.StartedAt.Format(time.RFC3339))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of executions to list")
	return cmd
}

func newExecGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [execution-id]",
		Short: "Get execution details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			executionID := args[0]

			execution, err := client.GetExecution(context.Background(), executionID)
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", execution.ID)
			fmt.Printf("Session ID: %s\n", execution.SessionID)
			fmt.Printf("Kind: %s\n", execution.Kind)
			fmt.Printf("Status: %s\n", execution.Status)
			fmt.Printf("Started: %s\n", execution.StartedAt.Format(time.RFC3339))

			if execution.EndedAt != nil {
				fmt.Printf("Ended: %s\n", execution.EndedAt.Format(time.RFC3339))
			}

			if execution.Input != "" {
				fmt.Printf("Input: %s\n", execution.Input)
			}

			if execution.Path != "" {
				fmt.Printf("Path: %s\n", execution.Path)
			}

			if execution.Result != nil {
				fmt.Printf("Result: %s\n", string(execution.Result))
			}

			if execution.Error != nil {
				fmt.Printf("Error: %s\n", string(execution.Error))
			}

			return nil
		},
	}
}

func newExecEventsCommand() *cobra.Command {
	var afterSeq int

	cmd := &cobra.Command{
		Use:   "events [execution-id]",
		Short: "Get execution events",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			executionID := args[0]

			events, err := client.GetExecutionEvents(context.Background(), executionID, afterSeq)
			if err != nil {
				return err
			}

			if len(events) == 0 {
				fmt.Println("No events found")
				return nil
			}

			fmt.Printf("%-5s %-20s %-15s %s\n", "Seq", "Timestamp", "Type", "Payload")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, event := range events {
				payloadStr := string(event.Payload)
				if len(payloadStr) > 50 {
					payloadStr = payloadStr[:47] + "..."
				}
				fmt.Printf("%-5d %-20s %-15s %s\n",
					event.Seq,
					event.Ts.Format("15:04:05"),
					event.Type,
					payloadStr)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&afterSeq, "after-seq", 0, "Get events after this sequence number")
	return cmd
}
