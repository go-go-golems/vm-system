package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/go-go-golems/vm-system/pkg/libloader"
	"github.com/go-go-golems/vm-system/pkg/vmmodels"
)

var libsCmd = &cobra.Command{
	Use:   "libs",
	Short: "Manage JavaScript libraries",
}

var downloadLibsCmd = &cobra.Command{
	Use:   "download",
	Short: "Download all builtin libraries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir := filepath.Join(".", ".vm-cache", "libraries")
		
		cache, err := libloader.NewLibraryCache(cacheDir)
		if err != nil {
			return err
		}
		
		fmt.Println("Downloading libraries to:", cacheDir)
		if err := cache.DownloadAll(); err != nil {
			return err
		}
		
		fmt.Println("\nCache info:")
		info := cache.GetCacheInfo()
		for id, cacheInfo := range info {
			fmt.Printf("  %s: %d bytes (%s)\n", id, cacheInfo.Size, cacheInfo.Path)
		}
		
		return nil
	},
}

var listLibsCmd = &cobra.Command{
	Use:   "list",
	Short: "List available libraries",
	RunE: func(cmd *cobra.Command, args []string) error {
		libraries := vmmodels.BuiltinLibraries()
		
		fmt.Printf("Available libraries (%d):\n\n", len(libraries))
		
		for _, lib := range libraries {
			fmt.Printf("  %s - %s v%s\n", lib.ID, lib.Name, lib.Version)
			fmt.Printf("    %s\n", lib.Description)
			fmt.Printf("    Source: %s\n", lib.Source)
			fmt.Printf("    Global: %s\n\n", lib.Config["global"])
		}
		
		return nil
	},
}

var cacheInfoCmd = &cobra.Command{
	Use:   "cache-info",
	Short: "Show library cache information",
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir := filepath.Join(".", ".vm-cache", "libraries")
		
		cache, err := libloader.NewLibraryCache(cacheDir)
		if err != nil {
			return err
		}
		
		// Load existing cache
		if err := cache.LoadExistingCache(); err != nil {
			return err
		}
		
		info := cache.GetCacheInfo()
		
		if len(info) == 0 {
			fmt.Println("No libraries cached. Run 'vm-system libs download' first.")
			return nil
		}
		
		fmt.Printf("Cached libraries (%d):\n\n", len(info))
		
		var totalSize int64
		for id, cacheInfo := range info {
			fmt.Printf("  %s\n", id)
			fmt.Printf("    Path: %s\n", cacheInfo.Path)
			fmt.Printf("    Size: %d bytes\n", cacheInfo.Size)
			fmt.Printf("    Modified: %s\n\n", cacheInfo.ModifiedTime.Format("2006-01-02 15:04:05"))
			totalSize += cacheInfo.Size
		}
		
		fmt.Printf("Total cache size: %d bytes (%.2f KB)\n", totalSize, float64(totalSize)/1024)
		
		return nil
	},
}

func init() {
	libsCmd.AddCommand(downloadLibsCmd)
	libsCmd.AddCommand(listLibsCmd)
	libsCmd.AddCommand(cacheInfoCmd)
}
