package vshnpostgres

import (
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	"github.com/go-logr/logr"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

// AddUserAlerting adds user alerting to the PostgreSQL instance.
func AddUserAlerting(log logr.Logger, iof *runtime.Runtime[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]) runtime.Result {

	log.Info("Check if alerting references are set")

	log.V(1).Info("Transforming", "obj", iof)

	err := runtime.AddToScheme(alertmanagerv1alpha1.SchemeBuilder)
	if err != nil {
		return runtime.NewFatal(err.Error())
	}

	monitoringSpec := comp.Spec.Parameters.Monitoring

	if monitoringSpec.AlertmanagerConfigRef != "" {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			log.Info("Found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef")
			return runtime.NewFatal("found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef, please specify as well")
		}

		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef
		log.Info("Found an AlertmanagerConfigRef, deploying...", "refName", refName)

		err = deployAlertmanagerFromRef(&iof.Desired.Composite, iof)
		if err != nil {
			return runtime.NewFatal(err.Error())
		}
	}

	if monitoringSpec.AlertmanagerConfigSpecTemplate != nil {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			log.Info("Found AlertmanagerConfigTemplate but no AlertmanagerConfigSecretRef")
			return runtime.NewFatal("found AlertmanagerConfigTemplate but no AlertmanagerConfigSecretRef, please specify as well")
		}

		log.Info("Found an AlertmanagerConfigTemplate, deploying...")

		err = deployAlertmanagerFromTemplate(comp, iof)
		if err != nil {
			return runtime.NewFatal(err.Error())
		}
	}

	if monitoringSpec.AlertmanagerConfigSecretRef != "" {
		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef
		log.Info("Found an AlertmanagerConfigSecretRef, deploying...", "refName", refName)

		err = deploySecretRef(comp, iof)
		if err != nil {
			return runtime.NewFatal(err.Error())
		}
	}

	return runtime.NewNormal()
}

func deployAlertmanagerFromRef(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]) error {
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

	return iof.Desired.PutIntoKubeObject(ac, comp.Name+"-alertmanagerconfig", xRef)
}

func deployAlertmanagerFromTemplate(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]) error {
	ac := &alertmanagerv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef,
			Namespace: comp.Status.InstanceNamespace,
		},
		Spec: *comp.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate,
	}

	return iof.Desired.PutIntoKubeObject(ac, comp.Name+"-alertmanagerconfig")
}

func deploySecretRef(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime[vshnv1.VSHNPostgreSQL, *vshnv1.VSHNPostgreSQL]) error {
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

	return iof.Desired.PutIntoKubeObject(s, comp.Name+"-alertmanagerconfigsecret", xRef)
}
