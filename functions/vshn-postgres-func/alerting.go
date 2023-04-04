package vshnpostgres

import (
	"context"
	"encoding/json"
	"fmt"

	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/go-logr/logr"
	alertmanagerv1alpha1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AddUserAlerting adds user alerting to the PostgreSQL instance.
func AddUserAlerting(ctx context.Context, log logr.Logger, iof *runtime.Runtime, comp *vshnv1.VSHNPostgreSQL) (*vshnv1.VSHNPostgreSQL, error) {

	log.Info("Check if alerting references are set")

	log.V(1).Info("Tranfsorming", "obj", iof)

	err := runtime.AddToScheme(alertmanagerv1alpha1.SchemeBuilder)
	if err != nil {
		return comp, err
	}

	monitoringSpec := comp.Spec.Parameters.Monitoring

	if monitoringSpec.AlertmanagerConfigRef != "" {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			log.Info("Found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef")
			return comp, fmt.Errorf("found AlertmanagerConfigRef but no AlertmanagerConfigSecretRef, please specify as well")
		}

		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigRef
		log.Info("Found an AlertmanagerConfigRef, deploying...", "refName", refName)

		err := deployAlertmanagerFromRef(comp, iof)
		if err != nil {
			return comp, err
		}

	}

	if monitoringSpec.AlertmanagerConfigSpecTemplate != nil {

		if monitoringSpec.AlertmanagerConfigSecretRef == "" {
			log.Info("Found AlertmanagerConfigTemplate but no AlertmanagerConfigSecretRef")
			return comp, fmt.Errorf("found AlertmanagerConfigTemplate but no AlertmanagerConfigSecretRef, please specify as well")
		}

		log.Info("Found an AlertmanagerConfigTemplate, deploying...")

		err := deployAlertmanagerFromTemplate(comp, iof)
		if err != nil {
			return comp, err
		}
	}

	if monitoringSpec.AlertmanagerConfigSecretRef != "" {
		refName := comp.Spec.Parameters.Monitoring.AlertmanagerConfigSecretRef
		log.Info("Found an AlertmanagerConfigSecretRef, deploying...", "refName", refName)

		err := deploySecretRef(comp, iof)
		if err != nil {
			return comp, err
		}
	}

	return comp, nil
}

func deployAlertmanagerFromRef(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	ac := &alertmanagerv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-alertmanagerconfig",
			Namespace: comp.Status.InstanceNamespace,
		},
	}

	xkobj, err := addObjectToXKube(comp, "-alertmanagerconfig", ac)
	if err != nil {
		return err
	}

	xkobj.Spec.References = []xkube.Reference{
		{
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
		},
	}

	err = iof.PutManagedRessource(xkobj)
	if err != nil {
		return err
	}
	return nil
}

func deployAlertmanagerFromTemplate(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	ac := &alertmanagerv1alpha1.AlertmanagerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-alertmanagerconfig",
			Namespace: comp.Status.InstanceNamespace,
		},
		Spec: *comp.Spec.Parameters.Monitoring.AlertmanagerConfigSpecTemplate,
	}

	xkobj, err := addObjectToXKube(comp, "-alertmanagerconfig", ac)
	if err != nil {
		return err
	}

	return iof.PutManagedRessource(xkobj)
}

func deploySecretRef(comp *vshnv1.VSHNPostgreSQL, iof *runtime.Runtime) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "postgresql-alertmanagerconfigsecret",
			Namespace: comp.Status.InstanceNamespace,
		},
	}
	xkobj, err := addObjectToXKube(comp, "-alertmanagerconfigsecret", secret)
	if err != nil {
		return err
	}

	xkobj.ObjectMeta.Name = comp.Name + "-alertmanagerconfigsecret"
	xkobj.Spec.References = []xkube.Reference{
		{
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
		},
	}

	return iof.PutManagedRessource(xkobj)
}

func addObjectToXKube(comp *vshnv1.VSHNPostgreSQL, namesuffix string, obj client.Object) (*xkube.Object, error) {

	err := runtime.SetGroupVersionKind(obj)
	if err != nil {
		return nil, err
	}

	rawData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	xkobj := &xkube.Object{
		ObjectMeta: metav1.ObjectMeta{
			Name: comp.Name + namesuffix,
		},
		Spec: xkube.ObjectSpec{
			ForProvider: xkube.ObjectParameters{
				Manifest: k8sruntime.RawExtension{
					Raw: rawData,
				},
			},
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "kubernetes",
				},
			},
		},
	}

	return xkobj, nil
}
