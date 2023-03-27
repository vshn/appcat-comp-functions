package src

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var S = runtime.NewScheme()

func init() {
	corev1.SchemeBuilder.AddToScheme(S)
	xkube.SchemeBuilder.SchemeBuilder.AddToScheme(S)
	vshnv1.SchemeBuilder.SchemeBuilder.AddToScheme(S)
}

var ErrNotFound = errors.New("not found")

// IO a struct which encapsulates crossplane FunctionIO and necessary k8s schemas
type IO struct {
	F xfnv1alpha1.FunctionIO
}

// NewFunctionIO creates a new IO object.
func NewFunctionIO(ctx context.Context) (*IO, error) {
	log := controllerruntime.LoggerFrom(ctx)

	funcIO := xfnv1alpha1.FunctionIO{}

	s := runtime.NewScheme()
	corev1.SchemeBuilder.AddToScheme(s)
	xkube.SchemeBuilder.SchemeBuilder.AddToScheme(s)
	vshnv1.SchemeBuilder.SchemeBuilder.AddToScheme(s)

	log.V(1).Info("Reading from stdin")
	x, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot read from stdin: %w", err)
	}

	log.V(1).Info("Unmarshalling FunctionIO from stdin")
	err = yaml.Unmarshal(x, &funcIO)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshall function io: %w", err)
	}
	return &IO{
		F: funcIO,
	}, nil
}

func (in *IO) GetFromKubeObject(resource client.Object, kubeObjectName string) error {
	ko := &xkube.Object{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeObjectName,
		},
	}
	err := in.get(ko)
	if err != nil {
		return fmt.Errorf("cannot get unmarshall kubernetes object: %w", err)
	}

	return in.fromKubeObject(ko, resource)
}

func (in *IO) fromKubeObject(kobj *xkube.Object, obj client.Object) error {
	if kobj.Spec.ForProvider.Manifest.Raw == nil {
		return fmt.Errorf("no resource in kubernetes object")
	}
	return json.Unmarshal(kobj.Spec.ForProvider.Manifest.Raw, obj)
}

func (in *IO) PutIntoKubeObject(resource client.Object, kubeObjectName string) error {
	ko, err := in.updateKubeObject(resource, kubeObjectName)
	if err != nil {
		return err
	}
	return in.put(ko)
}

func (in *IO) put(obj client.Object) error {
	name := obj.GetName()
	kind, _, err := S.ObjectKinds(obj)
	if err != nil {
		return err
	}

	obj.GetObjectKind().SetGroupVersionKind(kind[0])
	rawData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	for i, res := range in.F.Desired.Resources {
		if res.Name == name {
			in.F.Desired.Resources[i].Resource.Raw = rawData
			return nil
		}
	}

	in.F.Desired.Resources = append(in.F.Desired.Resources, xfnv1alpha1.DesiredResource{
		Name: name,
		Resource: runtime.RawExtension{
			Raw: rawData,
		},
	})
	return nil
}

func (in *IO) updateKubeObject(obj client.Object, kubeObjectName string) (*xkube.Object, error) {
	kind, _, err := S.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	obj.GetObjectKind().SetGroupVersionKind(kind[0])

	rawData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return &xkube.Object{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeObjectName,
		},
		Spec: xkube.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "kubernetes",
				},
			},
			ForProvider: xkube.ObjectParameters{Manifest: runtime.RawExtension{
				Raw: rawData,
			}},
		},
	}, nil
}

func (in *IO) get(obj client.Object) error {
	name := obj.GetName()
	for i, res := range in.F.Desired.Resources {
		if res.Name == name {
			return yaml.Unmarshal(in.F.Desired.Resources[i].Resource.Raw, obj)
		}
	}
	return ErrNotFound
}
