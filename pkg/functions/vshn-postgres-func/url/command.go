package url

import (
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-comp-functions/pkg"
	"github.com/vshn/appcat-comp-functions/pkg/functions/vshn-postgres-func"
)

type vshnPostgresURL struct{}

func NewVshnPostgresURL() *cli.Command {
	command := &vshnPostgresURL{}
	return &cli.Command{
		Name:   "vshn-postgres-url",
		Usage:  "Start VSHN Postgres URL Function IO",
		Action: command.execute,
	}
}

func (c *vshnPostgresURL) execute(ctx *cli.Context) error {
	_ = pkg.LogMetadata(ctx, vshnpostgres.AI)
	log := logr.FromContextOrDiscard(ctx.Context).WithName(ctx.Command.Name)
	log.Info("Executing FunctionIO - VSHN Postgres URL", "config", c)
	return pkg.Exec(ctx.Context, transform)
}
