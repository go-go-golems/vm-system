package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

type execReplSettings struct {
	SessionID string `glazed:"session-id"`
	Code      string `glazed:"code"`
}

type execRunFileSettings struct {
	SessionID string `glazed:"session-id"`
	Path      string `glazed:"path"`
	ArgsJSON  string `glazed:"args"`
	EnvJSON   string `glazed:"env"`
}

type execListSettings struct {
	SessionID string `glazed:"session-id"`
	Limit     int    `glazed:"limit"`
}

type execGetSettings struct {
	ExecutionID string `glazed:"execution-id"`
}

type execEventsSettings struct {
	ExecutionID string `glazed:"execution-id"`
	AfterSeq    int    `glazed:"after-seq"`
}

const (
	execActionRepl    = "repl"
	execActionRunFile = "run-file"
	execActionList    = "list"
	execActionGet     = "get"
	execActionEvents  = "events"
)

type execCommand struct {
	*cmds.CommandDescription
	action string
}

var _ cmds.WriterCommand = &execCommand{}

func (c *execCommand) RunIntoWriter(_ context.Context, vals *values.Values, w io.Writer) error {
	client := vmclient.New(serverURL, nil)

	switch c.action {
	case execActionRepl:
		settings := &execReplSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		execution, err := client.ExecuteREPL(context.Background(), vmclient.ExecuteREPLRequest{
			SessionID: settings.SessionID,
			Input:     settings.Code,
		})
		if err != nil {
			return err
		}

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
	case execActionRunFile:
		settings := &execRunFileSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		var argsMap map[string]interface{}
		var envMap map[string]interface{}
		if settings.ArgsJSON != "" {
			if err := json.Unmarshal([]byte(settings.ArgsJSON), &argsMap); err != nil {
				return fmt.Errorf("invalid args JSON: %w", err)
			}
		}
		if settings.EnvJSON != "" {
			if err := json.Unmarshal([]byte(settings.EnvJSON), &envMap); err != nil {
				return fmt.Errorf("invalid env JSON: %w", err)
			}
		}

		execution, err := client.ExecuteRunFile(context.Background(), vmclient.ExecuteRunFileRequest{
			SessionID: settings.SessionID,
			Path:      settings.Path,
			Args:      argsMap,
			Env:       envMap,
		})
		if err != nil {
			return err
		}

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
	case execActionList:
		settings := &execListSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		executions, err := client.ListExecutions(context.Background(), settings.SessionID, settings.Limit)
		if err != nil {
			return err
		}

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
	case execActionGet:
		settings := &execGetSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		execution, err := client.GetExecution(context.Background(), settings.ExecutionID)
		if err != nil {
			return err
		}

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
	case execActionEvents:
		settings := &execEventsSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		events, err := client.GetExecutionEvents(context.Background(), settings.ExecutionID, settings.AfterSeq)
		if err != nil {
			return err
		}

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
	default:
		return fmt.Errorf("unknown exec action: %s", c.action)
	}
}

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
	command := &execCommand{
		CommandDescription: commandDescription(
			"repl",
			"Execute REPL code",
			"Execute REPL code in a running session.",
			nil,
			[]*fields.Definition{
				fields.New("session-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Session ID")),
				fields.New("code", fields.TypeString, fields.WithRequired(true), fields.WithHelp("REPL code")),
			},
			false,
		),
		action: execActionRepl,
	}

	return buildCobraCommand(command)
}

func newExecRunFileCommand() *cobra.Command {
	command := &execCommand{
		CommandDescription: commandDescription(
			"run-file",
			"Run a file",
			"Execute a file path within a running session worktree.",
			[]*fields.Definition{
				fields.New("args", fields.TypeString, fields.WithDefault("{}"), fields.WithHelp("Arguments as JSON")),
				fields.New("env", fields.TypeString, fields.WithDefault("{}"), fields.WithHelp("Environment as JSON")),
			},
			[]*fields.Definition{
				fields.New("session-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Session ID")),
				fields.New("path", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Run-file path")),
			},
			false,
		),
		action: execActionRunFile,
	}

	return buildCobraCommand(command)
}

func newExecListCommand() *cobra.Command {
	command := &execCommand{
		CommandDescription: commandDescription(
			"list",
			"List executions for a session",
			"List executions for a session ID.",
			[]*fields.Definition{
				fields.New("limit", fields.TypeInteger, fields.WithDefault(50), fields.WithHelp("Maximum number of executions to list")),
			},
			[]*fields.Definition{
				fields.New("session-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Session ID")),
			},
			false,
		),
		action: execActionList,
	}

	return buildCobraCommand(command)
}

func newExecGetCommand() *cobra.Command {
	command := &execCommand{
		CommandDescription: commandDescription(
			"get",
			"Get execution details",
			"Get execution details by execution ID.",
			nil,
			[]*fields.Definition{
				fields.New("execution-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Execution ID")),
			},
			false,
		),
		action: execActionGet,
	}

	return buildCobraCommand(command)
}

func newExecEventsCommand() *cobra.Command {
	command := &execCommand{
		CommandDescription: commandDescription(
			"events",
			"Get execution events",
			"Get execution events by execution ID.",
			[]*fields.Definition{
				fields.New("after-seq", fields.TypeInteger, fields.WithDefault(0), fields.WithHelp("Get events after this sequence number")),
			},
			[]*fields.Definition{
				fields.New("execution-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Execution ID")),
			},
			false,
		),
		action: execActionEvents,
	}

	return buildCobraCommand(command)
}
