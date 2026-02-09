package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/vm-system/pkg/doc"
	"github.com/spf13/cobra"
)

var (
	dbPath    string
	serverURL string
)

func newRootCommand(helpSystem *help.HelpSystem) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "vm-system",
		Short: "JavaScript VM system with goja",
		Long:  `A VM subsystem that manages JavaScript execution with goja, integrating with dual-storage workspaces.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
	_ = logging.AddLoggingSectionToRootCommand(rootCmd, "vm-system")

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "vm-system.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "http://127.0.0.1:3210", "Daemon base URL for client mode commands")

	rootCmd.AddCommand(
		newServeCommand(),
		newTemplateCommand(),
		newSessionCommand(),
		newExecCommand(),
		newOpsCommand(),
		libsCmd,
	)

	return rootCmd
}

func main() {
	helpSystem := help.NewHelpSystem()
	_ = doc.AddDocToHelpSystem(helpSystem)
	rootCmd := newRootCommand(helpSystem)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
