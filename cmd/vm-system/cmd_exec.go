package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
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
		Use:   "repl <session-id> <code>",
		Short: "Execute REPL code",
		Long:  "Execute REPL code in a running session.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			code := args[1]

			client := vmclient.New(serverURL, nil)
			execution, err := client.ExecuteREPL(context.Background(), vmclient.ExecuteREPLRequest{
				SessionID: sessionID,
				Input:     code,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Execution ID: %s\n", execution.ID)
			_, _ = fmt.Fprintf(w, "Status: %s\n", execution.Status)
			if execution.Result != nil {
				_, _ = fmt.Fprintf(w, "Result: %s\n", string(execution.Result))
			}
			if execution.Error != nil {
				_, _ = fmt.Fprintf(w, "Error: %s\n", string(execution.Error))
			}

			events, err := client.GetExecutionEvents(context.Background(), execution.ID, 0)
			if err != nil {
				return err
			}

			if len(events) > 0 {
				_, _ = fmt.Fprintln(w, "\nEvents:")
				for _, event := range events {
					_, _ = fmt.Fprintf(w, "[%d] %s: %s\n", event.Seq, event.Type, string(event.Payload))
				}
			}

			return nil
		},
	}
}

func newExecRunFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-file <session-id> <path>",
		Short: "Run a file",
		Long:  "Execute a file path within a running session worktree.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			path := args[1]

			argsJSON, err := cmd.Flags().GetString("args")
			if err != nil {
				return err
			}
			envJSON, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}

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

			client := vmclient.New(serverURL, nil)
			execution, err := client.ExecuteRunFile(context.Background(), vmclient.ExecuteRunFileRequest{
				SessionID: sessionID,
				Path:      path,
				Args:      argsMap,
				Env:       envMap,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Execution ID: %s\n", execution.ID)
			_, _ = fmt.Fprintf(w, "Status: %s\n", execution.Status)
			if execution.Error != nil {
				_, _ = fmt.Fprintf(w, "Error: %s\n", string(execution.Error))
			}

			events, err := client.GetExecutionEvents(context.Background(), execution.ID, 0)
			if err != nil {
				return err
			}
			if len(events) > 0 {
				_, _ = fmt.Fprintln(w, "\nEvents:")
				for _, event := range events {
					_, _ = fmt.Fprintf(w, "[%d] %s: %s\n", event.Seq, event.Type, string(event.Payload))
				}
			}

			return nil
		},
	}

	cmd.Flags().String("args", "{}", "Arguments as JSON")
	cmd.Flags().String("env", "{}", "Environment as JSON")

	return cmd
}

func newExecListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <session-id>",
		Short: "List executions for a session",
		Long:  "List executions for a session ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			executions, err := client.ListExecutions(context.Background(), sessionID, limit)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if len(executions) == 0 {
				_, _ = fmt.Fprintln(w, "No executions found")
				return nil
			}

			_, _ = fmt.Fprintf(w, "%-36s %-10s %-10s %-20s\n", "Execution ID", "Kind", "Status", "Started")
			_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
			for _, execution := range executions {
				_, _ = fmt.Fprintf(w, "%-36s %-10s %-10s %-20s\n", execution.ID, execution.Kind, execution.Status, execution.StartedAt.Format(time.RFC3339))
			}
			return nil
		},
	}

	cmd.Flags().Int("limit", 50, "Maximum number of executions to list")

	return cmd
}

func newExecGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <execution-id>",
		Short: "Get execution details",
		Long:  "Get execution details by execution ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]

			client := vmclient.New(serverURL, nil)
			execution, err := client.GetExecution(context.Background(), executionID)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Execution ID: %s\n", execution.ID)
			_, _ = fmt.Fprintf(w, "Session ID: %s\n", execution.SessionID)
			_, _ = fmt.Fprintf(w, "Kind: %s\n", execution.Kind)
			_, _ = fmt.Fprintf(w, "Status: %s\n", execution.Status)
			_, _ = fmt.Fprintf(w, "Started: %s\n", execution.StartedAt.Format(time.RFC3339))
			if execution.EndedAt != nil {
				_, _ = fmt.Fprintf(w, "Ended: %s\n", execution.EndedAt.Format(time.RFC3339))
			}
			if execution.Input != "" {
				_, _ = fmt.Fprintf(w, "Input: %s\n", execution.Input)
			}
			if execution.Path != "" {
				_, _ = fmt.Fprintf(w, "Path: %s\n", execution.Path)
			}
			if execution.Result != nil {
				_, _ = fmt.Fprintf(w, "Result: %s\n", string(execution.Result))
			}
			if execution.Error != nil {
				_, _ = fmt.Fprintf(w, "Error: %s\n", string(execution.Error))
			}
			return nil
		},
	}
}

func newExecEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events <execution-id>",
		Short: "Get execution events",
		Long:  "Get execution events by execution ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executionID := args[0]
			afterSeq, err := cmd.Flags().GetInt("after-seq")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			events, err := client.GetExecutionEvents(context.Background(), executionID, afterSeq)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if len(events) == 0 {
				_, _ = fmt.Fprintln(w, "No events found")
				return nil
			}

			_, _ = fmt.Fprintf(w, "%-5s %-20s %-15s %s\n", "Seq", "Timestamp", "Type", "Payload")
			_, _ = fmt.Fprintln(w, "--------------------------------------------------------------------------------")
			for _, event := range events {
				payloadStr := string(event.Payload)
				if len(payloadStr) > 50 {
					payloadStr = payloadStr[:47] + "..."
				}
				_, _ = fmt.Fprintf(w, "%-5d %-20s %-15s %s\n", event.Seq, event.Ts.Format("15:04:05"), event.Type, payloadStr)
			}
			return nil
		},
	}

	cmd.Flags().Int("after-seq", 0, "Get events after this sequence number")

	return cmd
}
