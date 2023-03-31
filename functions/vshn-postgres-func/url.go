package vshnpostgres

import (
	"context"
	"fmt"
	"github.com/vshn/appcat-comp-functions/runtime"

	"github.com/go-logr/logr"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
)

// Transform changes the desired state of a FunctionIO
func Transform(ctx context.Context, log logr.Logger, iof *runtime.Runtime, comp *vshnv1.VSHNPostgreSQL) (*vshnv1.VSHNPostgreSQL, error) {

	fmt.Println("I'm a dummy!")

	return comp, nil
}
