package main

import "github.com/spf13/cobra"

func newHTTPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "HTTP API client commands",
		Long:  `Invoke daemon-backed HTTP API surfaces for templates, sessions, and executions.`,
	}

	cmd.AddCommand(
		newTemplateCommand(),
		newSessionCommand(),
		newExecCommand(),
	)

	return cmd
}
