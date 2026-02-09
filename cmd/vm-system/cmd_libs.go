package main

import (
	"fmt"
	"path/filepath"

	"github.com/go-go-golems/vm-system/pkg/libloader"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
	"github.com/spf13/cobra"
)

var libsCmd = newLibsCommand()

func newLibsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "libs",
		Short: "Manage JavaScript libraries",
	}

	cmd.AddCommand(
		newLibsDownloadCommand(),
		newLibsListCommand(),
		newLibsCacheInfoCommand(),
	)

	return cmd
}

func newLibsDownloadCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "download",
		Short: "Download all builtin libraries",
		Long:  "Download all built-in libraries into the local .vm-cache/libraries directory.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cacheDir := filepath.Join(".", ".vm-cache", "libraries")

			cache, err := libloader.NewLibraryCache(cacheDir)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Downloading libraries to: %s\n", cacheDir)
			if err := cache.DownloadAll(); err != nil {
				return err
			}

			_, _ = fmt.Fprintln(w, "\nCache info:")
			info := cache.GetCacheInfo()
			for id, cacheInfo := range info {
				_, _ = fmt.Fprintf(w, "  %s: %d bytes (%s)\n", id, cacheInfo.Size, cacheInfo.Path)
			}

			return nil
		},
	}
}

func newLibsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available libraries",
		Long:  "List available built-in JavaScript libraries.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			libraries := vmmodels.BuiltinLibraries()
			w := cmd.OutOrStdout()

			_, _ = fmt.Fprintf(w, "Available libraries (%d):\n\n", len(libraries))
			for _, lib := range libraries {
				_, _ = fmt.Fprintf(w, "  %s - %s v%s\n", lib.ID, lib.Name, lib.Version)
				_, _ = fmt.Fprintf(w, "    %s\n", lib.Description)
				_, _ = fmt.Fprintf(w, "    Source: %s\n", lib.Source)
				_, _ = fmt.Fprintf(w, "    Global: %s\n\n", lib.Config["global"])
			}

			return nil
		},
	}
}

func newLibsCacheInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cache-info",
		Short: "Show library cache information",
		Long:  "Show cached library file metadata and total cache size.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cacheDir := filepath.Join(".", ".vm-cache", "libraries")

			cache, err := libloader.NewLibraryCache(cacheDir)
			if err != nil {
				return err
			}

			if err := cache.LoadExistingCache(); err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			info := cache.GetCacheInfo()
			if len(info) == 0 {
				_, _ = fmt.Fprintln(w, "No libraries cached. Run 'vm-system libs download' first.")
				return nil
			}

			_, _ = fmt.Fprintf(w, "Cached libraries (%d):\n\n", len(info))

			var totalSize int64
			for id, cacheInfo := range info {
				_, _ = fmt.Fprintf(w, "  %s\n", id)
				_, _ = fmt.Fprintf(w, "    Path: %s\n", cacheInfo.Path)
				_, _ = fmt.Fprintf(w, "    Size: %d bytes\n", cacheInfo.Size)
				_, _ = fmt.Fprintf(w, "    Modified: %s\n\n", cacheInfo.ModifiedTime.Format("2006-01-02 15:04:05"))
				totalSize += cacheInfo.Size
			}

			_, _ = fmt.Fprintf(w, "Total cache size: %d bytes (%.2f KB)\n", totalSize, float64(totalSize)/1024)
			return nil
		},
	}
}
