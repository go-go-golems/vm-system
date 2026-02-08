package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/vm-system/pkg/vmclient"
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
		newSessionDeleteCommand(),
	)

	return cmd
}

func newSessionCreateCommand() *cobra.Command {
	var templateID, workspaceID, baseCommitOID, worktreePath string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VM session",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			fmt.Printf("Created session: %s\n", session.ID)
			fmt.Printf("Status: %s\n", session.Status)
			fmt.Printf("VM ID: %s\n", session.VMID)
			fmt.Printf("Workspace ID: %s\n", session.WorkspaceID)
			fmt.Printf("Base Commit: %s\n", session.BaseCommitOID)
			fmt.Printf("Worktree Path: %s\n", session.WorktreePath)

			return nil
		},
	}

	cmd.Flags().StringVar(&templateID, "template-id", "", "Template ID (required)")
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Workspace ID (required)")
	cmd.Flags().StringVar(&baseCommitOID, "base-commit", "", "Base commit OID (required)")
	cmd.Flags().StringVar(&worktreePath, "worktree-path", "", "Worktree path (required)")
	_ = cmd.MarkFlagRequired("template-id")
	_ = cmd.MarkFlagRequired("workspace-id")
	_ = cmd.MarkFlagRequired("base-commit")
	_ = cmd.MarkFlagRequired("worktree-path")

	return cmd
}

func newSessionListCommand() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VM sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessions, err := client.ListSessions(context.Background(), status)
			if err != nil {
				return err
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found")
				return nil
			}

			fmt.Printf("%-36s %-36s %-10s %-20s\n", "Session ID", "VM ID", "Status", "Created")
			fmt.Println("------------------------------------------------------------------------------------------------------")
			for _, session := range sessions {
				fmt.Printf("%-36s %-36s %-10s %-20s\n",
					session.ID,
					session.VMID,
					session.Status,
					session.CreatedAt.Format(time.RFC3339))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (starting, ready, crashed, closed)")
	return cmd
}

func newSessionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [session-id]",
		Short: "Get session details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessionID := args[0]

			session, err := client.GetSession(context.Background(), sessionID)
			if err != nil {
				return err
			}

			fmt.Printf("Session ID: %s\n", session.ID)
			fmt.Printf("VM ID: %s\n", session.VMID)
			fmt.Printf("Workspace ID: %s\n", session.WorkspaceID)
			fmt.Printf("Base Commit: %s\n", session.BaseCommitOID)
			fmt.Printf("Worktree Path: %s\n", session.WorktreePath)
			fmt.Printf("Status: %s\n", session.Status)
			fmt.Printf("Created: %s\n", session.CreatedAt.Format(time.RFC3339))

			if session.ClosedAt != nil {
				fmt.Printf("Closed: %s\n", session.ClosedAt.Format(time.RFC3339))
			}

			if session.LastError != "" {
				fmt.Printf("Last Error: %s\n", session.LastError)
			}

			return nil
		},
	}
}

func newSessionDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Close a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := vmclient.New(serverURL, nil)
			sessionID := args[0]
			if _, err := client.CloseSession(context.Background(), sessionID); err != nil {
				return err
			}

			fmt.Printf("Closed session: %s\n", sessionID)
			return nil
		},
	}
}
