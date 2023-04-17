package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var s = runtime.NewScheme()

type contextKey int

// Runtime a struct which encapsulates crossplane FunctionIO
type Runtime[T any, O interface {
	client.Object
	*T
}] struct {
	io       xfnv1alpha1.FunctionIO
	Observed ObservedResources[T, O]
	Desired  DesiredResources[T, O]
}

type Resource interface {
	GetName() string
	GetRaw() []byte
	SetRaw([]byte)
}

// KeyFuncIO is the key to the context value where the functionIO pointer is stored
const KeyFuncIO contextKey = iota

func init() {
	_ = corev1.SchemeBuilder.AddToScheme(s)
	_ = xkube.SchemeBuilder.SchemeBuilder.AddToScheme(s)
	_ = vshnv1.SchemeBuilder.SchemeBuilder.AddToScheme(s)
}

var ErrNotFound = errors.New("not found")

// NewRuntime creates a new Runtime object.
func NewRuntime[T any, O interface {
	client.Object
	*T
}](ctx context.Context) (*Runtime[T, O], error) {
	log := controllerruntime.LoggerFrom(ctx)

	log.V(1).Info("Reading from stdin")
	x, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("cannot read from stdin: %w", err)
	}

	log.V(1).Info("Unmarshalling FunctionIO from stdin")
	r := Runtime[T, O]{}
	err = yaml.Unmarshal(x, &r.io)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal function io: %w", err)
	}
	r.Observed = ObservedResources[T, O]{Resources: *observedResources(r.io.Observed.Resources)}
	r.Desired = DesiredResources[T, O]{Resources: *desiredResources(r.io.Desired.Resources)}

	log.V(1).Info("Unmarshalling observed composite from FunctionIO")
	var o T
	observed := &o
	err = json.Unmarshal(r.io.Observed.Composite.Resource.Raw, observed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal composite: %w", err)
	}
	r.Observed.Composite = *observed

	log.V(1).Info("Unmarshalling desired composite from FunctionIO")
	var d T
	desired := &d
	err = json.Unmarshal(r.io.Observed.Composite.Resource.Raw, desired)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal composite: %w", err)
	}
	r.Desired.Composite = *desired

	return &r, nil
}

func fromKubeObject(kobj *xkube.Object, obj client.Object) error {
	if kobj.Status.AtProvider.Manifest.Raw == nil {
		if kobj.Spec.ForProvider.Manifest.Raw == nil {
			return fmt.Errorf("no resource in kubernetes object")
		}
		return json.Unmarshal(kobj.Spec.ForProvider.Manifest.Raw, obj)
	}
	return json.Unmarshal(kobj.Status.AtProvider.Manifest.Raw, obj)
}

func getKubeObjectFrom(ctx context.Context, resources *[]Resource, o client.Object, kon string) (*xkube.Object, error) {
	log := controllerruntime.LoggerFrom(ctx)

	ko := &xkube.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       xkube.ObjectKind,
			APIVersion: xkube.ObjectKindAPIVersion,
		},
	}
	err := getFrom(resources, ko, kon)
	if err != nil {
		return nil, fmt.Errorf("cannot get unmarshall kubernetes object: %w", err)
	}

	log.V(1).Info("Unmarshalling object from kube object", "object type", reflect.TypeOf(o))
	return ko, nil
}

func getFrom(resources *[]Resource, obj client.Object, resName string) error {
	gvk := obj.GetObjectKind()

	for _, res := range *resources {
		if res.GetName() == resName {
			err := yaml.Unmarshal(res.GetRaw(), obj)
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

func desiredResources(dr []xfnv1alpha1.DesiredResource) *[]Resource {
	resources := make([]Resource, len(dr))

	for i := range dr {
		resources[i] = desiredResource(dr[i])
	}

	return &resources
}

func observedResources(or []xfnv1alpha1.ObservedResource) *[]Resource {
	resources := make([]Resource, len(or))

	for i := range or {
		resources[i] = observedResource(or[i])
	}

	return &resources
}

func updateKubeObject(obj client.Object, ko *xkube.Object) error {
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

// AddToScheme adds given SchemeBuilder to the Scheme.
func AddToScheme(obj runtime.SchemeBuilder) error {
	return obj.AddToScheme(s)
}
