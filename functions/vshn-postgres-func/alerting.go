package vshnpostgres

import (
	"context"
	controllerruntime "sigs.k8s.io/controller-runtime"

	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// AddUserAlerting adds user alerting to the PostgreSQL instance.
func AddUserAlerting(ctx context.Context, iof *runtime.Runtime) runtime.Result {
	log := controllerruntime.LoggerFrom(ctx)
	log.Info("Check if alerting references are set")

	log.V(1).Info("Transforming", "obj", iof)

	err := runtime.AddToScheme(alertmanagerv1alpha1.SchemeBuilder)
	if err != nil {
		return runtime.NewFatalErr(ctx, "Cannot add scheme builder to scheme", err)
	}
	comp := &vshnv1.VSHNPostgreSQL{}
	err = iof.Observed.GetComposite(ctx, comp)
	if err != nil {
		return runtime.NewFatalErr(ctx, "Cannot get composite from function io", err)
	}

	monitoringSpec := comp.Spec.Parameters.Monitoring

	if monitoringSpec.AlertmanagerConfigRef != "" {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			return runtime.NewFatal(ctx, "Found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef, please specify as well")
		}

		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef
		log.Info("Found an AlertmanagerConfigRef, deploying...", "refName", refName)

		err = deployAlertmanagerFromRef(ctx, comp, iof)
		if err != nil {
			return runtime.NewFatalErr(ctx, "Could not deploy alertmanager from ref", err)
		}
	}

	if monitoringSpec.AlertmanagerConfigSpecTemplate != nil {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			return runtime.NewFatal(ctx, "Found AlertmanagerConfigTemplate but no AlertmanagerConfigSecretRef, please specify as well")
		}

		log.Info("Found an AlertmanagerConfigTemplate, deploying...")

		err = deployAlertmanagerFromTemplate(ctx, comp, iof)
		if err != nil {
			return runtime.NewFatalErr(ctx, "Cannot deploy alertmanager from template", err)
		}
	}

	if monitoringSpec.AlertmanagerConfigSecretRef != "" {
		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef
		log.Info("Found an AlertmanagerConfigSecretRef, deploying...", "refName", refName)

		err = deploySecretRef(ctx, comp, iof)
		if err != nil {
			return runtime.NewFatalErr(ctx, "Cannot deploy secret ref", err)
		}
	}
	/*
		err = iof.Desired.SetComposite(ctx, comp)
		if err != nil {
			return runtime.NewFatalErr(ctx, "Cannot set desired composite", err)
		}
	*/
	return runtime.NewNormal()
}

func deployAlertmanagerFromRef(ctx context.Context, comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	ac := &alertmanagerv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-alertmanagerconfig",
			Namespace: comp.Status.InstanceNamespace,
		},
	}

	xRef := xkube.Reference{
		PatchesFrom: &xkube.PatchesFrom{
			DependsOn: xkube.DependsOn{
				APIVersion: "monitoring.coreos.com/v1alpha1",
				Kind:       "AlertmanagerConfig",
				Namespace:  comp.ObjectMeta.Labels["crossplane.io/claim-namespace"],
				Name:       comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef,
			},
			FieldPath: pointer.String("spec"),
		},
		ToFieldPath: pointer.String("spec"),
	}

	return iof.Desired.PutIntoKubeObject(ctx, ac, comp.Name+"-alertmanagerconfig", xRef)
}

func deployAlertmanagerFromTemplate(ctx context.Context, comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	ac := &alertmanagerv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef,
			Namespace: comp.Status.InstanceNamespace,
		},
		Spec: *comp.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate,
	}

	return iof.Desired.PutIntoKubeObject(ctx, ac, comp.Name+"-alertmanagerconfig")
}

func deploySecretRef(ctx context.Context, comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef,
			Namespace: comp.Status.InstanceNamespace,
		},
	}
	xRef := xkube.Reference{
		PatchesFrom: &xkube.PatchesFrom{
			DependsOn: xkube.DependsOn{
				APIVersion: "v1",
				Kind:       "Secret",
				Namespace:  comp.ObjectMeta.Labels["crossplane.io/claim-namespace"],
				Name:       comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef,
			},
			FieldPath: pointer.String("data"),
		},
		ToFieldPath: pointer.String("data"),
	}

	return iof.Desired.PutIntoKubeObject(ctx, s, comp.Name+"-alertmanagerconfigsecret", xRef)
}
