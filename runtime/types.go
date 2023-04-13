package runtime

import (
	"github.com/go-logr/logr"
)

// AppInfo defines application information
type AppInfo struct {
	Version, Commit, Date, AppName, AppLongName string
}

// Transform specifies a transformation function to be run against the given FunctionIO.
type Transform struct {
	Name          string
	TransformFunc func(log logr.Logger, io *Runtime[T, O]) Result
}
