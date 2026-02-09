package main

import "github.com/spf13/cobra"

type templateCreateSettings struct {
	Name   string `glazed:"name"`
	Engine string `glazed:"engine"`
}

type templateIDArg struct {
	TemplateID string `glazed:"template-id"`
}

type templateNameFlag struct {
	TemplateID string `glazed:"template-id"`
	Name       string `glazed:"name"`
}

type templateAddCapabilitySettings struct {
	TemplateID string `glazed:"template-id"`
	Kind       string `glazed:"kind"`
	Name       string `glazed:"name"`
	ConfigJSON string `glazed:"config"`
	Enabled    bool   `glazed:"enabled"`
}

type templateAddStartupSettings struct {
	TemplateID string `glazed:"template-id"`
	Path       string `glazed:"path"`
	Mode       string `glazed:"mode"`
	OrderIndex int    `glazed:"order"`
}

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage templates via daemon API",
		Long:  `Create, list, inspect, and delete templates plus capability/startup policy metadata through the daemon REST API.`,
	}

	cmd.AddCommand(
		newTemplateCreateCommand(),
		newTemplateListCommand(),
		newTemplateGetCommand(),
		newTemplateDeleteCommand(),
		newTemplateAddModuleCommand(),
		newTemplateRemoveModuleCommand(),
		newTemplateListModulesCommand(),
		newTemplateAddLibraryCommand(),
		newTemplateRemoveLibraryCommand(),
		newTemplateListLibrariesCommand(),
		newTemplateListAvailableModulesCommand(),
		newTemplateListAvailableLibrariesCommand(),
		newTemplateAddCapabilityCommand(),
		newTemplateListCapabilitiesCommand(),
		newTemplateAddStartupFileCommand(),
		newTemplateListStartupFilesCommand(),
	)

	return cmd
}
