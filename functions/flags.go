package functions

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func init() {
	// Remove `-v` short option from --version flag in favor of verbosity.
	cli.VersionFlag.(*cli.BoolFlag).Aliases = nil
}

func NewLogLevelFlag() *cli.IntFlag {
	return &cli.IntFlag{
		Name: "log-level", Aliases: []string{"v"}, EnvVars: []string{"LOG_LEVEL"},
		Usage: "number of the log level verbosity",
		Value: 0,
	}
}

func NewLogFormatFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name: "log-format", EnvVars: []string{"LOG_FORMAT"},
		Usage: "sets the log format (values: [json, console])",
		Value: "console",
		Action: func(context *cli.Context, format string) error {
			if format == "console" || format == "json" {
				return nil
			}
			_ = cli.ShowAppHelp(context)
			return fmt.Errorf("unknown log format: %s", format)
		},
	}
}
