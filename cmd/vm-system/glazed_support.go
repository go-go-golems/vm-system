package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
)

func buildCobraCommand(c cmds.Command) *cobra.Command {
	cobraCmd, err := cli.BuildCobraCommand(
		c,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		panic(err)
	}
	return cobraCmd
}

func commandDescription(name, short, long string, flags, args []*fields.Definition, withOutput bool) *cmds.CommandDescription {
	sections := []schema.Section{}
	if withOutput {
		glazedSection, err := settings.NewGlazedSchema()
		if err != nil {
			panic(err)
		}
		sections = append(sections, glazedSection)
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		panic(err)
	}
	sections = append(sections, commandSettingsSection)

	opts := []cmds.CommandDescriptionOption{
		cmds.WithShort(short),
		cmds.WithLong(long),
		cmds.WithSections(sections...),
	}
	if len(flags) > 0 {
		opts = append(opts, cmds.WithFlags(flags...))
	}
	if len(args) > 0 {
		opts = append(opts, cmds.WithArguments(args...))
	}

	return cmds.NewCommandDescription(name, opts...)
}

func decodeDefault(vals *values.Values, dst interface{}) error {
	return vals.DecodeSectionInto(schema.DefaultSlug, dst)
}
