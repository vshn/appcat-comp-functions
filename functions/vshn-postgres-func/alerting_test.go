package vshnpostgres

import (
	"context"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	v1 "k8s.io/api/core/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddUserAlerting(t *testing.T) {
	ctx := context.Background()

	type args struct {
		expectedFuncIO string
		inputFuncIO    string
	}
	tests := []struct {
		name      string
		args      args
		expResult xfnv1alpha1.Result
	}{
		{
			name: "GivenNoMonitoringParams_ThenExpectNoOutput",
			args: args{
				expectedFuncIO: "alerting/01-ThenExpectNoOutput.yaml",
				inputFuncIO:    "alerting/01-GivenNoMonitoringParams.yaml",
			},
			expResult: xfnv1alpha1.Result{
				Severity: xfnv1alpha1.SeverityNormal,
				Message:  fmt.Sprintf("function ran successfully"),
			},
		},
		{
			name:      "GivenConfigRefNoSecretRef_ThenExpectError",
			expResult: runtime.NewFatal(ctx, "Found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef, please specify as well").Resolve(),
			args: args{
				expectedFuncIO: "alerting/02-ThenExpectError.yaml",
				inputFuncIO:    "alerting/02-GivenConfigRefNoSecretRef.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			iof := loadRuntimeFromFile(t, tt.args.inputFuncIO)
			expIof := loadRuntimeFromFile(t, tt.args.expectedFuncIO)

			r := AddUserAlerting(ctx, iof)

			assert.Equal(t, tt.expResult, r.Resolve())
			assert.Equal(t, getFunctionIo(expIof), getFunctionIo(iof))
		})
	}
}

func TestGivenConfigRefAndSecretThenExpectOutput(t *testing.T) {

	ctx := context.Background()

	t.Run("GivenConfigRefAndSecret_ThenExpectOutput", func(t *testing.T) {

		iof := loadRuntimeFromFile(t, "alerting/03-GivenConfigRefAndSecret.yaml")

		r := AddUserAlerting(ctx, iof)
		assert.Equal(t, runtime.NewNormal(), r)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.Desired.GetManagedResource(resName, kubeObject))

		assert.Equal(t, iof.Desired.Composite.Labels["crossplane.io/claim-namespace"], kubeObject.Spec.References[0].PatchesFrom.Namespace)
		assert.Equal(t, iof.Desired.Composite.Spec.Parameters.Monitoring.AlertmanagerConfigRef, kubeObject.Spec.References[0].PatchesFrom.Name)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.Desired.GetFromKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, iof.Desired.Composite.Status.InstanceNamespace, alertConfig.GetNamespace())

		secretName := "psql-alertmanagerconfigsecret"
		secret := &v1.Secret{}
		assert.NoError(t, iof.Desired.GetFromKubeObject(ctx, secret, secretName))

		assert.Equal(t, iof.Desired.Composite.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef, secret.GetName())
	})
}

func TestGivenConfigTemplateAndSecretThenExpectOutput(t *testing.T) {
	ctx := context.Background()

	t.Run("GivenConfigTemplateAndSecret_ThenExpectOutput", func(t *testing.T) {

		iof := loadRuntimeFromFile(t, "alerting/04-GivenConfigTemplateAndSecret.yaml")

		r := AddUserAlerting(ctx, iof)
		assert.Equal(t, runtime.NewNormal(), r)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.Desired.GetManagedResource(resName, kubeObject))

		assert.Empty(t, kubeObject.Spec.References)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.Desired.GetFromKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, iof.Desired.Composite.Status.InstanceNamespace, alertConfig.GetNamespace())
		assert.Equal(t, iof.Desired.Composite.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate, &alertConfig.Spec)

		secretName := "psql-alertmanagerconfigsecret"
		secret := &v1.Secret{}
		assert.NoError(t, iof.Desired.GetFromKubeObject(ctx, secret, secretName))

		assert.Equal(t, iof.Desired.Composite.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef, secret.GetName())
	})
}
