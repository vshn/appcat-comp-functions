package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	vp "github.com/vshn/appcat-comp-functions/functions/vshn-postgres-func"
	"github.com/vshn/appcat-comp-functions/runtime"
	"os"

	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
)

var postgresFunctions = []runtime.Transform[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]{
	{
		Name:          "url-connection-detail",
		TransformFunc: vp.Transform,
	},
}

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
		Action:  run,
		Flags: []cli.Flag{
			runtime.NewLogLevelFlag(),
			runtime.NewLogFormatFlag(),
			runtime.NewFunctionFlag(),
		},
	}
	return app
}

func run(ctx *cli.Context) error {
	err := runtime.SetupLogging(vp.AI, ctx)
	if err != nil {
		return err
	}

	_ = runtime.LogMetadata(ctx, vp.AI)

	return runtime.RunCommand(ctx, postgresFunctions)
}
