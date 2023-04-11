package vshnpostgres

import (
	"context"
	"testing"

	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	"github.com/go-logr/logr"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/stretchr/testify/assert"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
)

func TestAddUserAlerting(t *testing.T) {
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
				expectedFuncIO: "alerting/01-ThenExpectNoOutput.yaml",
				inputFuncIO:    "alerting/01-GivenNoMonitoringParams.yaml",
			},
			wantErr: false,
		},
		{
			name:    "GivenConfigRefNoSecretRef_ThenExpectError",
			wantErr: true,
			args: args{
				expectedFuncIO: "alerting/02-ThenExpectError.yaml",
				inputFuncIO:    "alerting/02-GivenConfigRefNoSecretRef.yaml",
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

			_, err := AddUserAlerting(ctx, log, iof, inComp)

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

		iof := getFunctionFromFile(t, "alerting/03-GivenConfigRefAndSecret.yaml")
		comp := &vshnv1.VSHNPostgreSQL{}
		inComp := getCompositeFromIO(t, iof, comp)

		_, err := AddUserAlerting(ctx, log, iof, inComp)
		assert.NoError(t, err)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.GetManagedRessourceFromDesired(resName, kubeObject))

		assert.Equal(t, comp.Labels["crossplane.io/claim-namespace"], kubeObject.Spec.References[0].PatchesFrom.Namespace)
		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef, kubeObject.Spec.References[0].PatchesFrom.Name)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, comp.Status.InstanceNamespace, alertConfig.GetNamespace())

		secretName := "psql-alertmanagerconfigsecret"
		secret := &v1.Secret{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, secret, secretName))

		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef, secret.GetName())
	})

}

func TestGivenConfigTemplateAndSecretThenExpectOutput(t *testing.T) {
	ctx := context.Background()
	log := logr.FromContextOrDiscard(ctx)

	t.Run("GivenConfigTemplateAndSecret_ThenExpectOutput", func(t *testing.T) {

		iof := getFunctionFromFile(t, "alerting/04-GivenConfigTemplateAndSecret.yaml")
		comp := &vshnv1.VSHNPostgreSQL{}
		inComp := getCompositeFromIO(t, iof, comp)

		_, err := AddUserAlerting(ctx, log, iof, inComp)
		assert.NoError(t, err)

		resName := "psql-alertmanagerconfig"
		kubeObject := &xkube.Object{}
		assert.NoError(t, iof.GetManagedRessourceFromDesired(resName, kubeObject))

		assert.Empty(t, kubeObject.Spec.References)

		alertConfig := &alertmanagerv1alpha1.AlertmanagerConfig{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, alertConfig, resName))
		assert.Equal(t, comp.Status.InstanceNamespace, alertConfig.GetNamespace())
		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate, &alertConfig.Spec)

		secretName := "psql-alertmanagerconfigsecret"
		secret := &v1.Secret{}
		assert.NoError(t, iof.GetFromDesiredKubeObject(ctx, secret, secretName))

		assert.Equal(t, comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef, secret.GetName())
	})
}
