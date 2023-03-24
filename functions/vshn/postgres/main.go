package postgres

import (
	"encoding/json"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	"github.com/vshn/go-bootstrap/lib"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
)

var (
	PostgresqlHost     = "POSTGRESQL_HOST"
	PostgresqlUser     = "POSTGRESQL_USER"
	PostgresqlPassword = "POSTGRESQL_PASSWORD"
	PostgresqlPort     = "POSTGRESQL_PORT"
	PostgresqlDb       = "POSTGRESQL_DB"
	PostgresqlUrl      = "POSTGRESQL_URL"
)

// transform changes the desired state of FunctionIO
func transform(iof *lib.IO, comp *vshnv1.VSHNPostgreSQL) (*vshnv1.VSHNPostgreSQL, error) {
	// Wait for the next reconciliation in case instance namespace is missing
	if comp.Status.InstanceNamespace == "" {
		return comp, nil
	}

	so := &xkube.Object{
		ObjectMeta: metav1.ObjectMeta{
			//TODO check if namespace exists otherwise wait for next reconciliation
			Namespace: comp.Status.InstanceNamespace,
			Name:      comp.Name + "-connection",
		},
	}

	err := iof.GetKubeObj(so)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot get connection secret object: %v", err))
	}

	s := &v1.Secret{}
	err = json.Unmarshal(so.Spec.ForProvider.Manifest.Raw, s)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot get connection secret: %v", err))
	}

	return comp, addPostgresURL(s)
}

func addPostgresURL(s *v1.Secret) error {
	u := string(s.Data[PostgresqlUser])
	pwd := string(s.Data[PostgresqlPassword])
	h := string(s.Data[PostgresqlHost])
	p := string(s.Data[PostgresqlPort])
	db := string(s.Data[PostgresqlDb])

	// The values are still missing, wait for the next reconciliation
	if u == "" || pwd == "" || h == "" || p == "" || db == "" {
		return nil
	}

	if len(s.Data) == 0 {
		log.Fatal(fmt.Errorf("no data found in connection secret"))
	}

	s.Data[PostgresqlUrl] = []byte("postgres://" +
		string(s.Data[PostgresqlUser]) + ":" +
		string(s.Data[PostgresqlPassword]) + "@" +
		string(s.Data[PostgresqlHost]) + ":" +
		string(s.Data[PostgresqlPort]) + "/" +
		string(s.Data[PostgresqlDb]))

	return nil
}

func main() {
	sb := []runtime.SchemeBuilder{
		corev1.SchemeBuilder,
		xkube.SchemeBuilder.SchemeBuilder,
		vshnv1.SchemeBuilder.SchemeBuilder,
	}
	lib.Exec(transform, sb...)
}
