package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var s = runtime.NewScheme()

type contextKey int

// KeyFuncIO is the key to the context value where the functionIO pointer is stored
const KeyFuncIO contextKey = iota

func init() {
	_ = corev1.SchemeBuilder.AddToScheme(s)
	_ = xkube.SchemeBuilder.SchemeBuilder.AddToScheme(s)
	_ = vshnv1.SchemeBuilder.SchemeBuilder.AddToScheme(s)
}

var ErrNotFound = errors.New("not found")

// Runtime a struct which encapsulates crossplane FunctionIO
type Runtime xfnv1alpha1.FunctionIO

// getFunctionIO creates a new Runtime object.
func getFunctionIO(ctx context.Context) (*Runtime, error) {
	log := controllerruntime.LoggerFrom(ctx)

	log.V(1).Info("Reading from stdin")
	x, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot read from stdin: %w", err)
	}

	log.V(1).Info("Unmarshalling FunctionIO from stdin")
	funcIO := Runtime{}
	err = yaml.Unmarshal(x, &funcIO)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal function io: %w", err)
	}

	return &funcIO, nil
}

// AddToScheme adds given SchemeBuilder to the Scheme.
func AddToScheme(obj runtime.SchemeBuilder) error {
	return obj.AddToScheme(s)
}

// GetFromKubeObject gets the k8s resource o from a provider kubernetes object kon
func (in *Runtime) GetFromKubeObject(ctx context.Context, o client.Object, kon string) error {
	log := controllerruntime.LoggerFrom(ctx)

	log.V(1).Info("Creating kube object from name and unmarshalling it", "kube object", kon)
	ko := &xkube.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       xkube.ObjectKind,
			APIVersion: xkube.ObjectKindAPIVersion,
		},
	}
	err := in.get(ko, kon)
	if err != nil {
		return fmt.Errorf("cannot get unmarshal kubernetes object: %w", err)
	}

	log.V(1).Info("Unmarshalling object from kube object", "object type", reflect.TypeOf(o))
	return in.fromKubeObject(ko, o)
}

func (in *Runtime) fromKubeObject(kobj *xkube.Object, obj client.Object) error {
	if kobj.Status.AtProvider.Manifest.Raw == nil {
		return fmt.Errorf("no resource in kubernetes object")
	}
	return json.Unmarshal(kobj.Status.AtProvider.Manifest.Raw, obj)
}

// PutIntoKubeObject adds or updates the desired resource into its kube object
func (in *Runtime) PutIntoKubeObject(ctx context.Context, o client.Object, kon string) error {
	log := controllerruntime.LoggerFrom(ctx)

	ko := &xkube.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       xkube.ObjectKind,
			APIVersion: xkube.ObjectKindAPIVersion,
		},
	}
	err := in.get(ko, kon)
	if err != nil {
		return err
	}

	log.V(1).Info("Put object into kube object", "object", o, "kube object name", kon)
	err = in.updateKubeObject(o, ko)
	if err != nil {
		return err
	}
	return in.put(ko, kon)
}

func (in *Runtime) put(obj client.Object, resName string) error {
	kind, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}

	obj.GetObjectKind().SetGroupVersionKind(kind[0])
	rawData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	for i, res := range in.Desired.Resources {
		if res.Name == resName {
			in.Desired.Resources[i].Resource.Raw = rawData
			return nil
		}
	}

	in.Desired.Resources = append(in.Desired.Resources, xfnv1alpha1.DesiredResource{
		Name: resName,
		Resource: runtime.RawExtension{
			Raw: rawData,
		},
	})
	return nil
}

func (in *Runtime) updateKubeObject(obj client.Object, ko *xkube.Object) error {
	kind, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(kind[0])

	rawData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	ko.Spec.ForProvider.Manifest = runtime.RawExtension{Raw: rawData}

	return nil
}

func (in *Runtime) get(obj client.Object, resName string) error {
	gvk := obj.GetObjectKind()

	for i, res := range in.Observed.Resources {
		if res.Name == resName {
			err := yaml.Unmarshal(in.Observed.Resources[i].Resource.Raw, obj)
			if err != nil {
				return fmt.Errorf("cannot unmarshal desired resource: %w", err)
			}

			// matching by name is not enough, group and kind should match
			ogvk := obj.GetObjectKind()
			if gvk == ogvk {
				return nil
			}
		}
	}
	return ErrNotFound
}

// AddResult will add a new result to the results array.
// These results will generate events on the composite.
func (in *Runtime) AddResult(severity xfnv1alpha1.Severity, message string) {
	in.Results = append(in.Results, xfnv1alpha1.Result{
		Severity: severity,
		Message:  message,
	})
}

// PutManagedRessource will add the object as is to the FunctionIO. It assumes
// that the given object is adheres to Crossplane's ManagedResource model.
func (in *Runtime) PutManagedRessource(obj client.Object) error {
	return in.put(obj, obj.GetName())
}

// GetManagedRessource will unmarshall the managed resource with the given name into the
// given object.
func (in *Runtime) GetManagedRessource(resName string, obj client.Object) error {
	return in.get(obj, resName)
}

// SetGroupVersionKind automatically populates the GVK of an object with
// the current scheme.
func SetGroupVersionKind(obj client.Object) error {
	kind, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(kind[0])
	return nil
}
