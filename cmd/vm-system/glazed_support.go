package main

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func buildCobraCommand(c cmds.Command) *cobra.Command {
	cobraCmd, err := buildCobraCommandE(c)
	if err != nil {
		name := c.Description().Name
		log.Error().
			Err(err).
			Str("command", name).
			Msg("failed to build cobra command")
		return &cobra.Command{
			Use:           name,
			Short:         c.Description().Short,
			SilenceUsage:  true,
			SilenceErrors: true,
			RunE: func(_ *cobra.Command, _ []string) error {
				return fmt.Errorf("failed to initialize command %q: %w", name, err)
			},
		}
	}
	return cobraCmd
}

func buildCobraCommandE(c cmds.Command) (*cobra.Command, error) {
	cobraCmd, err := cli.BuildCobraCommand(
		c,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		return nil, err
	}
	return cobraCmd, nil
}

func commandDescription(name, short, long string, flags, args []*fields.Definition, withOutput bool) *cmds.CommandDescription {
	description, err := commandDescriptionE(name, short, long, flags, args, withOutput)
	if err != nil {
		log.Error().
			Err(err).
			Str("command", name).
			Msg("failed to create full command description, using fallback schema")
		opts := []cmds.CommandDescriptionOption{
			cmds.WithShort(short),
			cmds.WithLong(long),
		}
		if len(flags) > 0 {
			opts = append(opts, cmds.WithFlags(flags...))
		}
		if len(args) > 0 {
			opts = append(opts, cmds.WithArguments(args...))
		}
		return cmds.NewCommandDescription(name, opts...)
	}
	return description
}

func commandDescriptionE(name, short, long string, flags, args []*fields.Definition, withOutput bool) (*cmds.CommandDescription, error) {
	sections := []schema.Section{}
	if withOutput {
		glazedSection, err := settings.NewGlazedSchema()
		if err != nil {
			return nil, fmt.Errorf("create glazed schema: %w", err)
		}
		sections = append(sections, glazedSection)
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, fmt.Errorf("create command settings section: %w", err)
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

	return cmds.NewCommandDescription(name, opts...), nil
}

func decodeDefault(vals *values.Values, dst interface{}) error {
	return vals.DecodeSectionInto(schema.DefaultSlug, dst)
}
