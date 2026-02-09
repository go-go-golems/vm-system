package main

import "github.com/spf13/cobra"

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
