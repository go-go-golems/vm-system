package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmexec"
	"github.com/go-go-golems/vm-system/pkg/vmsession"
	"github.com/go-go-golems/vm-system/pkg/vmstore"
)

var executor *vmexec.Executor

func newExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute code in VM sessions",
		Long:  `Execute REPL snippets or run files in VM sessions.`,
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
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if sessionManager == nil {
				sessionManager = getSessionManager(store)
			}
			if executor == nil {
				executor = vmexec.NewExecutor(store, sessionManager)
			}

			sessionID := args[0]
			code := args[1]

			exec, err := executor.ExecuteREPL(sessionID, code)
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", exec.ID)
			fmt.Printf("Status: %s\n", exec.Status)
			
			if exec.Result != nil {
				fmt.Printf("Result: %s\n", string(exec.Result))
			}
			
			if exec.Error != nil {
				fmt.Printf("Error: %s\n", string(exec.Error))
			}

			// Show events
			events, err := executor.GetEvents(exec.ID, 0)
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
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if sessionManager == nil {
				sessionManager = getSessionManager(store)
			}
			if executor == nil {
				executor = vmexec.NewExecutor(store, sessionManager)
			}

			sessionID := args[0]
			path := args[1]

			// Parse args and env
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

			exec, err := executor.ExecuteRunFile(sessionID, path, argsMap, envMap)
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", exec.ID)
			fmt.Printf("Status: %s\n", exec.Status)
			
			if exec.Error != nil {
				fmt.Printf("Error: %s\n", string(exec.Error))
			}

			// Show events
			events, err := executor.GetEvents(exec.ID, 0)
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
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if sessionManager == nil {
				sessionManager = getSessionManager(store)
			}
			if executor == nil {
				executor = vmexec.NewExecutor(store, sessionManager)
			}

			sessionID := args[0]

			executions, err := executor.ListExecutions(sessionID, limit)
			if err != nil {
				return err
			}

			if len(executions) == 0 {
				fmt.Println("No executions found")
				return nil
			}

			fmt.Printf("%-36s %-10s %-10s %-20s\n", "Execution ID", "Kind", "Status", "Started")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, exec := range executions {
				fmt.Printf("%-36s %-10s %-10s %-20s\n",
					exec.ID,
					exec.Kind,
					exec.Status,
					exec.StartedAt.Format(time.RFC3339))
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
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if sessionManager == nil {
				sessionManager = getSessionManager(store)
			}
			if executor == nil {
				executor = vmexec.NewExecutor(store, sessionManager)
			}

			executionID := args[0]

			exec, err := executor.GetExecution(executionID)
			if err != nil {
				return err
			}

			fmt.Printf("Execution ID: %s\n", exec.ID)
			fmt.Printf("Session ID: %s\n", exec.SessionID)
			fmt.Printf("Kind: %s\n", exec.Kind)
			fmt.Printf("Status: %s\n", exec.Status)
			fmt.Printf("Started: %s\n", exec.StartedAt.Format(time.RFC3339))
			
			if exec.EndedAt != nil {
				fmt.Printf("Ended: %s\n", exec.EndedAt.Format(time.RFC3339))
			}
			
			if exec.Input != "" {
				fmt.Printf("Input: %s\n", exec.Input)
			}
			
			if exec.Path != "" {
				fmt.Printf("Path: %s\n", exec.Path)
			}
			
			if exec.Result != nil {
				fmt.Printf("Result: %s\n", string(exec.Result))
			}
			
			if exec.Error != nil {
				fmt.Printf("Error: %s\n", string(exec.Error))
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
			store, err := vmstore.NewVMStore(dbPath)
			if err != nil {
				return err
			}
			defer store.Close()

			if sessionManager == nil {
				sessionManager = getSessionManager(store)
			}
			if executor == nil {
				executor = vmexec.NewExecutor(store, sessionManager)
			}

			executionID := args[0]

			events, err := executor.GetEvents(executionID, afterSeq)
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

// Helper to get or create session manager
func getSessionManager(store *vmstore.VMStore) *vmsession.SessionManager {
	return vmsession.NewSessionManager(store)
}
