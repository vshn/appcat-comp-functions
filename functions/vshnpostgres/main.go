package main

import (
	"fmt"
	"github.com/vshn/appcat-comp-functions/functions"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var A = functions.AppInfo{
	Version:     "unknown",
	Commit:      "-dirty-",
	Date:        time.Now().Format("2006-01-02"),
	AppName:     "functionio-vshn-postgresql-url",
	AppLongName: "A crossplane composition function to craft an URK from an instance of a vshn postgres database",
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
		Name:    A.AppName,
		Usage:   A.AppLongName,
		Version: fmt.Sprintf("%s, revision=%s, date=%s", A.Version, A.Commit, A.Date),

		Before: functions.SetupLogging(A),
		Flags: []cli.Flag{
			functions.NewLogLevelFlag(),
			functions.NewLogFormatFlag(),
		},
		Commands: []*cli.Command{
			NewVshnPostgresURL(),
		},
	}
	return app
}
