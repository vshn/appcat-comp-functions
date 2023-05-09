package vshnpostgres

import (
	"fmt"

	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
)

func getInstanceNamespace(comp *vshnv1.VSHNPostgreSQL) string {
	return fmt.Sprintf("vshn-postgresql-%s", comp.GetName())
}
