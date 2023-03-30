//go:build integration

package vshnpostgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
)

func TestTransform_NoInstanceNamespace(t *testing.T) {
	expectIo := getFunctionFromFile("01_expected_no-instance-namespace.yaml")
	expectVpu := &vshnv1.VSHNPostgreSQL{}
	expectComp := getCompositeFromIO(expectIo, expectVpu)

	t.Run("WhenNoInstance_ThenNoErrorAndNoChanges", func(t *testing.T) {

		//Given
		io := getFunctionFromFile("01_input_no-instance-namespace.yaml")
		vpu := &vshnv1.VSHNPostgreSQL{}
		comp := getCompositeFromIO(io, vpu)
		ctx := context.Background()
		log := logr.FromContextOrDiscard(ctx)

		// When
		comp, err := Transform(ctx, log, io, comp)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectIo, io)
		assert.Equal(t, expectComp, comp)
	})
}

func TestTransform(t *testing.T) {
	expectURL := "postgres://postgres:639b-9076-4de6-a35@" +
		"pgsql-gc9x4.vshn-postgresql-pgsql-gc9x4.svc.cluster.local:5432/postgres"

	t.Run("WhenNormalIO_ThenAddPostgreSQLUrl", func(t *testing.T) {

		//Given
		io := getFunctionFromFile("02_input_function-io.yaml")
		vpu := &vshnv1.VSHNPostgreSQL{}
		comp := getCompositeFromIO(io, vpu)
		ctx := context.Background()
		log := logr.FromContextOrDiscard(ctx)

		// When
		actualComp, err := Transform(ctx, log, io, comp)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, expectURL, io.Desired.Composite.ConnectionDetails[0].Value)
		assert.Equal(t, comp, actualComp)
	})
}

func TestGetPostgresURL(t *testing.T) {
	tests := map[string]struct {
		secret    *v1.Secret
		expectUrl string
	}{
		"WhenMissingUserAndPortThenReturnNoUrlInSecret": {
			secret: &v1.Secret{
				Data: map[string][]byte{
					PostgresqlPassword: []byte("test"),
					PostgresqlDb:       []byte("db-test"),
					PostgresqlHost:     []byte("localhost"),
				},
			},
			expectUrl: "",
		},
		"WhenMissingPasswordThenReturnNoUrlInSecret": {
			secret: &v1.Secret{
				Data: map[string][]byte{
					PostgresqlDb:   []byte("db-test"),
					PostgresqlHost: []byte("localhost"),
					PostgresqlUser: []byte("user"),
					PostgresqlPort: []byte("5432"),
				},
			},
			expectUrl: "",
		},
		"WhenDataThenReturnSecretWithUrl": {
			secret: &v1.Secret{
				Data: map[string][]byte{
					PostgresqlPassword: []byte("test"),
					PostgresqlDb:       []byte("db-test"),
					PostgresqlHost:     []byte("localhost"),
					PostgresqlUser:     []byte("user"),
					PostgresqlPort:     []byte("5432"),
				},
				StringData: map[string]string{
					"place": "test",
				},
			},
			expectUrl: "postgres://user:test@localhost:5432/db-test",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// Given
			ctx := context.Background()

			// When
			url := getPostgresURL(ctx, tc.secret)

			// Then
			assert.Equal(t, tc.expectUrl, url)
		})
	}
}

func getFunctionFromFile(file string) *runtime.Runtime {
	p, _ := filepath.Abs(".")
	before, _, _ := strings.Cut(p, "/functions")
	dat, err := os.ReadFile(before + "/test/transforms/vshn-postgres/url/" + file)
	if err != nil {
		fmt.Errorf("cannot read test file %s: %w", file, err)
	}

	funcIO := runtime.Runtime{}
	err = yaml.Unmarshal(dat, &funcIO)
	if err != nil {
		fmt.Errorf("cannot umarshal test file %s: %w", file, err)
	}

	return &funcIO
}

func getCompositeFromIO[T any](io *runtime.Runtime, obj *T) *T {
	err := json.Unmarshal(io.Observed.Composite.Resource.Raw, obj)
	if err != nil {
		os.Exit(1)
	}
	return obj
}
