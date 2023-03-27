package main

import (
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-comp-functions/functions"
	"github.com/vshn/appcat-comp-functions/src"
)

type vshnPostgresURL struct{}

func NewVshnPostgresURL() *cli.Command {
	command := &vshnPostgresURL{}
	return &cli.Command{
		Name:   "vshn-postgres-url",
		Usage:  "Start VSHN Postgres URL",
		Action: command.execute,
	}
}

func (c *vshnPostgresURL) execute(ctx *cli.Context) error {
	_ = functions.LogMetadata(ctx, A)
	log := logr.FromContextOrDiscard(ctx.Context).WithName(ctx.Command.Name)
	log.Info("Executing FunctionIO VSHN Postgres URL", "config", c)
	return src.Exec(ctx.Context, transform)
}
