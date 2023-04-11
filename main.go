package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-comp-functions/controller"
	vp "github.com/vshn/appcat-comp-functions/functions/vshn-postgres-func"
	"github.com/vshn/appcat-comp-functions/runtime"

	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
)

var postgresFunctions = []runtime.Transform[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]{
	{
		Name:          "url-connection-detail",
		TransformFunc: vp.Transform,
	},
	{
		Name:          "user-alerting",
		TransformFunc: vp.AddUserAlerting,
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
		Before:  setupLogging,
		Flags: []cli.Flag{
			runtime.NewLogLevelFlag(),
			runtime.NewLogFormatFlag(),
			runtime.NewFunctionFlag(),
		},
		Commands: []*cli.Command{
			{
				Name:        "controller",
				Description: "Runs the controller mode of the composition function runner",
				Action:      controller.RunController,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "metrics-addr",
						Value: ":8080",
						Usage: "The address the metric endpoint binds to.",
					},
					&cli.StringFlag{
						Name:  "health-addr",
						Value: ":8081",
						Usage: "The address the probe endpoint binds to.",
					},
					&cli.BoolFlag{
						Name:  "leader-elect",
						Value: false,
						Usage: "Enable leader election for controller manager. " +
							"Enabling this will ensure there is only one active controller manager.",
					},
				},
			},
		},
	}
	return app
}

func run(ctx *cli.Context) error {

	return runtime.RunCommand(ctx, postgresFunctions)
}

func setupLogging(ctx *cli.Context) error {
	err := runtime.SetupLogging(vp.AI, ctx)
	if err != nil {
		return err
	}

	return runtime.LogMetadata(ctx, vp.AI)
}
