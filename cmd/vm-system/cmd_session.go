package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
	"github.com/spf13/cobra"
)

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
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VM session",
		Long:  "Create a new VM runtime session from a template.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			templateID, err := cmd.Flags().GetString("template-id")
			if err != nil {
				return err
			}
			workspaceID, err := cmd.Flags().GetString("workspace-id")
			if err != nil {
				return err
			}
			baseCommitOID, err := cmd.Flags().GetString("base-commit")
			if err != nil {
				return err
			}
			worktreePath, err := cmd.Flags().GetString("worktree-path")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			session, err := client.CreateSession(context.Background(), vmclient.CreateSessionRequest{
				TemplateID:    templateID,
				WorkspaceID:   workspaceID,
				BaseCommitOID: baseCommitOID,
				WorktreePath:  worktreePath,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Created session: %s\n", session.ID)
			_, _ = fmt.Fprintf(w, "Status: %s\n", session.Status)
			_, _ = fmt.Fprintf(w, "Template ID: %s\n", session.VMID)
			_, _ = fmt.Fprintf(w, "Workspace ID: %s\n", session.WorkspaceID)
			_, _ = fmt.Fprintf(w, "Base Commit: %s\n", session.BaseCommitOID)
			_, _ = fmt.Fprintf(w, "Worktree Path: %s\n", session.WorktreePath)
			return nil
		},
	}

	cmd.Flags().String("template-id", "", "Template ID (required)")
	cmd.Flags().String("workspace-id", "", "Workspace ID (required)")
	cmd.Flags().String("base-commit", "", "Base commit OID (required)")
	cmd.Flags().String("worktree-path", "", "Worktree path (required)")
	_ = cmd.MarkFlagRequired("template-id")
	_ = cmd.MarkFlagRequired("workspace-id")
	_ = cmd.MarkFlagRequired("base-commit")
	_ = cmd.MarkFlagRequired("worktree-path")

	return cmd
}

func newSessionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VM sessions",
		Long:  "List VM sessions and optionally filter by status.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := cmd.Flags().GetString("status")
			if err != nil {
				return err
			}

			client := vmclient.New(serverURL, nil)
			sessions, err := client.ListSessions(context.Background(), status)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
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
		},
	}

	cmd.Flags().String("status", "", "Filter by status (starting, ready, crashed, closed)")

	return cmd
}

func newSessionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id>",
		Short: "Get session details",
		Long:  "Get session details by session ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			client := vmclient.New(serverURL, nil)
			session, err := client.GetSession(context.Background(), sessionID)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
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
		},
	}
}

func newSessionCloseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "close <session-id>",
		Short: "Close a session",
		Long:  "Close a VM session by ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			client := vmclient.New(serverURL, nil)
			if _, err := client.CloseSession(context.Background(), sessionID); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Closed session: %s\n", sessionID)
			return nil
		},
	}
}
