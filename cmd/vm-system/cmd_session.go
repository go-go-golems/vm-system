package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

type sessionCreateSettings struct {
	TemplateID    string `glazed:"template-id"`
	WorkspaceID   string `glazed:"workspace-id"`
	BaseCommitOID string `glazed:"base-commit"`
	WorktreePath  string `glazed:"worktree-path"`
}

type sessionListSettings struct {
	Status string `glazed:"status"`
}

type sessionIDArg struct {
	SessionID string `glazed:"session-id"`
}

const (
	sessionActionCreate = "create"
	sessionActionList   = "list"
	sessionActionGet    = "get"
	sessionActionClose  = "close"
)

type sessionCommand struct {
	*cmds.CommandDescription
	action string
}

var _ cmds.WriterCommand = &sessionCommand{}

func (c *sessionCommand) RunIntoWriter(_ context.Context, vals *values.Values, w io.Writer) error {
	client := vmclient.New(serverURL, nil)

	switch c.action {
	case sessionActionCreate:
		settings := &sessionCreateSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		session, err := client.CreateSession(context.Background(), vmclient.CreateSessionRequest{
			TemplateID:    settings.TemplateID,
			WorkspaceID:   settings.WorkspaceID,
			BaseCommitOID: settings.BaseCommitOID,
			WorktreePath:  settings.WorktreePath,
		})
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Created session: %s\n", session.ID)
		_, _ = fmt.Fprintf(w, "Status: %s\n", session.Status)
		_, _ = fmt.Fprintf(w, "Template ID: %s\n", session.VMID)
		_, _ = fmt.Fprintf(w, "Workspace ID: %s\n", session.WorkspaceID)
		_, _ = fmt.Fprintf(w, "Base Commit: %s\n", session.BaseCommitOID)
		_, _ = fmt.Fprintf(w, "Worktree Path: %s\n", session.WorktreePath)
		return nil
	case sessionActionList:
		settings := &sessionListSettings{}
		if err := decodeDefault(vals, settings); err != nil {
			return err
		}

		sessions, err := client.ListSessions(context.Background(), settings.Status)
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			_, _ = fmt.Fprintln(w, "No sessions found")
			return nil
		}

		_, _ = fmt.Fprintf(w, "%-36s %-36s %-10s %-20s\n", "Session ID", "Template ID", "Status", "Created")
		_, _ = fmt.Fprintln(w, "------------------------------------------------------------------------------------------------------")
		for _, session := range sessions {
			_, _ = fmt.Fprintf(w, "%-36s %-36s %-10s %-20s\n", session.ID, session.VMID, session.Status, session.CreatedAt.Format(time.RFC3339))
		}
		return nil
	case sessionActionGet:
		args := &sessionIDArg{}
		if err := decodeDefault(vals, args); err != nil {
			return err
		}

		session, err := client.GetSession(context.Background(), args.SessionID)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Session ID: %s\n", session.ID)
		_, _ = fmt.Fprintf(w, "Template ID: %s\n", session.VMID)
		_, _ = fmt.Fprintf(w, "Workspace ID: %s\n", session.WorkspaceID)
		_, _ = fmt.Fprintf(w, "Base Commit: %s\n", session.BaseCommitOID)
		_, _ = fmt.Fprintf(w, "Worktree Path: %s\n", session.WorktreePath)
		_, _ = fmt.Fprintf(w, "Status: %s\n", session.Status)
		_, _ = fmt.Fprintf(w, "Created: %s\n", session.CreatedAt.Format(time.RFC3339))
		if session.ClosedAt != nil {
			_, _ = fmt.Fprintf(w, "Closed: %s\n", session.ClosedAt.Format(time.RFC3339))
		}
		if session.LastError != "" {
			_, _ = fmt.Fprintf(w, "Last Error: %s\n", session.LastError)
		}
		return nil
	case sessionActionClose:
		args := &sessionIDArg{}
		if err := decodeDefault(vals, args); err != nil {
			return err
		}

		if _, err := client.CloseSession(context.Background(), args.SessionID); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Closed session: %s\n", args.SessionID)
		return nil
	default:
		return fmt.Errorf("unknown session action: %s", c.action)
	}
}

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage VM sessions via daemon API",
		Long:  `Create, list, and manage VM runtime sessions through the daemon REST API.`,
	}

	cmd.AddCommand(
		newSessionCreateCommand(),
		newSessionListCommand(),
		newSessionGetCommand(),
		newSessionCloseCommand(),
	)

	return cmd
}

func newSessionCreateCommand() *cobra.Command {
	command := &sessionCommand{
		CommandDescription: commandDescription(
			"create",
			"Create a new VM session",
			"Create a new VM runtime session from a template.",
			[]*fields.Definition{
				fields.New("template-id", fields.TypeString, fields.WithHelp("Template ID (required)"), fields.WithRequired(true)),
				fields.New("workspace-id", fields.TypeString, fields.WithHelp("Workspace ID (required)"), fields.WithRequired(true)),
				fields.New("base-commit", fields.TypeString, fields.WithHelp("Base commit OID (required)"), fields.WithRequired(true)),
				fields.New("worktree-path", fields.TypeString, fields.WithHelp("Worktree path (required)"), fields.WithRequired(true)),
			},
			nil,
			false,
		),
		action: sessionActionCreate,
	}

	return buildCobraCommand(command)
}

func newSessionListCommand() *cobra.Command {
	command := &sessionCommand{
		CommandDescription: commandDescription(
			"list",
			"List VM sessions",
			"List VM sessions and optionally filter by status.",
			[]*fields.Definition{
				fields.New("status", fields.TypeString, fields.WithDefault(""), fields.WithHelp("Filter by status (starting, ready, crashed, closed)")),
			},
			nil,
			false,
		),
		action: sessionActionList,
	}

	return buildCobraCommand(command)
}

func newSessionGetCommand() *cobra.Command {
	command := &sessionCommand{
		CommandDescription: commandDescription(
			"get",
			"Get session details",
			"Get session details by session ID.",
			nil,
			[]*fields.Definition{
				fields.New("session-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Session ID")),
			},
			false,
		),
		action: sessionActionGet,
	}

	return buildCobraCommand(command)
}

func newSessionCloseCommand() *cobra.Command {
	command := &sessionCommand{
		CommandDescription: commandDescription(
			"close",
			"Close a session",
			"Close a VM session by ID.",
			nil,
			[]*fields.Definition{
				fields.New("session-id", fields.TypeString, fields.WithRequired(true), fields.WithHelp("Session ID")),
			},
			false,
		),
		action: sessionActionClose,
	}

	return buildCobraCommand(command)
}
