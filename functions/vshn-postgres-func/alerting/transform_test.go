package alerting

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	"github.com/go-logr/logr"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	"sigs.k8s.io/yaml"
)

func TestTransform(t *testing.T) {
	type args struct {
		expectedFuncIO string
		inputFuncIO    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "GivenNoMonitoringParams_ThenExpectNoOutput",
			args: args{
				expectedFuncIO: "01-ThenExpectNoOutput.yaml",
				inputFuncIO:    "01-GivenNoMonitoringParams.yaml",
			},
			wantErr: false,
		},
		{
			name:    "GivenConfigRefNoSecretRef_ThenExpectError",
			wantErr: true,
			args: args{
				expectedFuncIO: "02-ThenExpectError.yaml",
				inputFuncIO:    "02-GivenConfigRefNoSecretRef.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			log := logr.FromContextOrDiscard(ctx)

			iof := getFunctionFromFile(t, tt.args.inputFuncIO)
			comp := &vshnv1.VSHNPostgreSQL{}
			inComp := getCompositeFromIO(t, iof, comp)
			expIof := getFunctionFromFile(t, tt.args.expectedFuncIO)

			_, err := Transform(ctx, log, iof, inComp)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expIof, iof)
			}

		})
	}
}

func TestGivenConfigRefAndSecretThenExpectOutput(t *testing.T) {

	ctx := context.Background()
	log := logr.FromContextOrDiscard(ctx)

	t.Run("GivenConfigRefAndSecret_ThenExpectOutput", func(t *testing.T) {

		iof := getFunctionFromFile(t, "03-GivenConfigRefAndSecret.yaml")
		comp := &vshnv1.VSHNPostgreSQL{}
		inComp := getCompositeFromIO(t, iof, comp)

		_, err := Transform(ctx, log, iof, inComp)
		assert.NoError(t, err)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.GetManagedRessourceFromDesired(resName, kubeObject))

		assert.Equal(t, comp.Labels["crossplane.io/claim-namespace"], kubeObject.Spec.References[0].PatchesFrom.Namespace)
		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef, kubeObject.Spec.References[0].PatchesFrom.Name)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, comp.Status.InstanceNamespace, alertConfig.GetNamespace())
	})

}

func TestGivenConfigTemplateAndSecretThenExpectOutput(t *testing.T) {
	ctx := context.Background()
	log := logr.FromContextOrDiscard(ctx)

	t.Run("GivenConfigTemplateAndSecret_ThenExpectOutput", func(t *testing.T) {

		iof := getFunctionFromFile(t, "04-GivenConfigTemplateAndSecret.yaml")
		comp := &vshnv1.VSHNPostgreSQL{}
		inComp := getCompositeFromIO(t, iof, comp)

		_, err := Transform(ctx, log, iof, inComp)
		assert.NoError(t, err)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.GetManagedRessourceFromDesired(resName, kubeObject))

		assert.Empty(t, kubeObject.Spec.References)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, comp.Status.InstanceNamespace, alertConfig.GetNamespace())
		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate, &alertConfig.Spec)
	})
}

func getFunctionFromFile(t assert.TestingT, file string) *runtime.Runtime {
	p, _ := filepath.Abs(".")
	before, _, _ := strings.Cut(p, "/functions")
	dat, err := os.ReadFile(before + "/test/transforms/vshn-postgres/alerting/" + file)
	assert.NoError(t, err)

	funcIO := runtime.Runtime{}
	err = yaml.Unmarshal(dat, &funcIO)
	assert.NoError(t, err)

	return &funcIO
}

func getCompositeFromIO[T any](t assert.TestingT, io *runtime.Runtime, obj *T) *T {
	err := json.Unmarshal(io.Observed.Composite.Resource.Raw, obj)
	assert.NoError(t, err)

	return obj
}
