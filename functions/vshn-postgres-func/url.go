package vshnpostgres

import (
	"context"
	"github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var (
	// PostgresqlHost is env variable in the connection secret
	PostgresqlHost = "POSTGRESQL_HOST"
	// PostgresqlUser is env variable in the connection secret
	PostgresqlUser = "POSTGRESQL_USER"
	// PostgresqlPassword is env variable in the connection secret
	PostgresqlPassword = "POSTGRESQL_PASSWORD"
	// PostgresqlPort is env variable in the connection secret
	PostgresqlPort = "POSTGRESQL_PORT"
	// PostgresqlDb is env variable in the connection secret
	PostgresqlDb = "POSTGRESQL_DB"
	// PostgresqlUrl is env variable in the connection secret
	PostgresqlUrl = "POSTGRESQL_URL"
)

// connectionSecretResourceName is the resource name defined in the composition
// This name is different from metadata.name of the same resource
// The value is hardcoded in the composition for each resource and due to crossplane limitation
// it cannot be matched to the metadata.name
var connectionSecretResourceName = "connection"

// AddUrlToConnectionDetails changes the desired state of a FunctionIO
func AddUrlToConnectionDetails(ctx context.Context, iof *runtime.Runtime[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]) runtime.Result {
	log := controllerruntime.LoggerFrom(ctx)

	// Wait for the next reconciliation in case instance namespace is missing
	if iof.Observed.Composite.Status.InstanceNamespace == "" {
		return runtime.NewWarning(ctx, "Composite is missing instance namespace, skipping transformation")
	}

	log.Info("Getting connection secret from managed kubernetes object")
	s := &v1.Secret{}

	err := iof.Observed.GetFromKubeObject(ctx, s, connectionSecretResourceName)
	if err != nil {
		return runtime.NewFatalErr(ctx, "Cannot get connection secret object", err)
	}

	log.Info("Setting POSTRESQL_URL env variable into connection secret")
	val := getPostgresURL(s)

	iof.Desired.ConnectionDetails =
		append(iof.Desired.ConnectionDetails, v1alpha1.ExplicitConnectionDetail{
			Name:  PostgresqlUrl,
			Value: val,
		})

	if val == "" {
		return runtime.NewWarning(ctx, "User, pass, host, port or db value is missing from connection secret, skipping transformation")
	}

	return runtime.NewNormal()
}

func getPostgresURL(s *v1.Secret) string {
	user := string(s.Data[PostgresqlUser])
	pwd := string(s.Data[PostgresqlPassword])
	host := string(s.Data[PostgresqlHost])
	port := string(s.Data[PostgresqlPort])
	db := string(s.Data[PostgresqlDb])

	// The values are still missing, wait for the next reconciliation
	if user == "" || pwd == "" || host == "" || port == "" || db == "" {
		return ""
	}

	return "postgres://" + user + ":" + pwd + "@" + host + ":" + port + "/" + db
}
