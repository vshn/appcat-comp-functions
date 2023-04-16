package runtime

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AppInfo defines application information
type AppInfo struct {
	Version, Commit, Date, AppName, AppLongName string
}

// Transform specifies a transformation function to be run against the given FunctionIO.
type Transform[T any, O interface {
	client.Object
	*T
}] struct {
	Name          string
	TransformFunc func(c context.Context, io *Runtime[T, O]) Result
}
