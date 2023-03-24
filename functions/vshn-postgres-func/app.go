package vshnpostgres

import (
	"github.com/vshn/appcat-comp-functions/runtime"
	"time"
)

var AI = runtime.AppInfo{
	Version:     "unknown",
	Commit:      "-dirty-",
	Date:        time.Now().Format("2006-01-02"),
	AppName:     "functionio-vshn-postgresql-url",
	AppLongName: "A crossplane composition function to craft an URL from an instance of a vshn postgres database",
}
