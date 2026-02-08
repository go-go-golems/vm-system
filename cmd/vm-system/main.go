package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dbPath    string
	serverURL string
	rootCmd   *cobra.Command
)

func main() {
	rootCmd = &cobra.Command{
		Use:   "vm-system",
		Short: "JavaScript VM system with goja",
		Long:  `A VM subsystem that manages JavaScript execution with goja, integrating with dual-storage workspaces.`,
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "vm-system.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server-url", "http://127.0.0.1:3210", "Daemon base URL for client mode commands")

	rootCmd.AddCommand(
		newServeCommand(),
		newTemplateCommand(),
		newSessionCommand(),
		newExecCommand(),
		modulesCmd,
		libsCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
