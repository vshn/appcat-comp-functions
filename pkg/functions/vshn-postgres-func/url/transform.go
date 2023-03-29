package url

import (
	"context"
	"fmt"
	"github.com/vshn/appcat-comp-functions/pkg"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var (
	PostgresqlHost     = "POSTGRESQL_HOST"
	PostgresqlUser     = "POSTGRESQL_USER"
	PostgresqlPassword = "POSTGRESQL_PASSWORD"
	PostgresqlPort     = "POSTGRESQL_PORT"
	PostgresqlDb       = "POSTGRESQL_DB"
	PostgresqlUrl      = "POSTGRESQL_URL"
)

// transform changes the desired state of a FunctionIO
func transform(ctx context.Context, iof *pkg.IO, comp *vshnv1.VSHNPostgreSQL) (*vshnv1.VSHNPostgreSQL, error) {
	log := controllerruntime.LoggerFrom(ctx)

	// Wait for the next reconciliation in case instance namespace is missing
	if comp.Status.InstanceNamespace == "" {
		log.Info("Composite is missing instance namespace, skipping transformation")
		return comp, nil
	}

	log.Info("Getting connection secret from managed kubernetes object")
	s := &v1.Secret{}
	kon := comp.Name + "-connection"
	err := iof.GetFromKubeObject(ctx, s, kon)
	if err != nil {
		return nil, fmt.Errorf("cannot get connection secret object: %w", err)
	}

	log.Info("Setting POSTRESQL_URL env variable into connection secret")
	err = addPostgresURL(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("cannot update secret with postgres url: %w", err)
	}

	log.Info("Updating desired FunctionIO state")
	err = iof.PutIntoKubeObject(ctx, s, kon)
	if err != nil {
		return nil, err
	}

	return comp, nil
}

func addPostgresURL(ctx context.Context, s *v1.Secret) error {
	log := controllerruntime.LoggerFrom(ctx)

	user := s.StringData[PostgresqlUser]
	pwd := string(s.Data[PostgresqlPassword])
	host := s.StringData[PostgresqlHost]
	port := s.StringData[PostgresqlPort]
	db := s.StringData[PostgresqlDb]

	// The values are still missing, wait for the next reconciliation
	if user == "" || pwd == "" || host == "" || port == "" || db == "" {
		log.Info("User, pass, host, port or db value is missing from connection secret, skipping transformation")
		return nil
	}

	s.Data[PostgresqlUrl] = []byte("postgres://" + user + ":" + pwd + "@" + host + ":" + port + "/" + db)
	return nil
}
