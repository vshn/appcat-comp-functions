package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-comp-functions/pkg"
	vp "github.com/vshn/appcat-comp-functions/pkg/functions/vshn-postgres"
	"github.com/vshn/appcat-comp-functions/pkg/functions/vshn-postgres/url"
	"os"
)

func main() {
	app := newApp()
	err := app.Run(os.Args)
	// If required flags aren't set, it will return with error before we could set up logging
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	app := &cli.App{
		Name:    vp.AI.AppName,
		Usage:   vp.AI.AppLongName,
		Version: fmt.Sprintf("%s, revision=%s, date=%s", vp.AI.Version, vp.AI.Commit, vp.AI.Date),

		Before: pkg.SetupLogging(vp.AI),
		Flags: []cli.Flag{
			pkg.NewLogLevelFlag(),
			pkg.NewLogFormatFlag(),
		},
		Commands: []*cli.Command{
			url.NewVshnPostgresURL(),
		},
	}
	return app
}
