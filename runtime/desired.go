package runtime

import (
	"context"
	"encoding/json"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredResources[T any, O interface {
	client.Object
	*T
}] struct {
	Resources         []Resource
	Composite         T
	ConnectionDetails []xfnv1alpha1.ExplicitConnectionDetail
}

// GetFromKubeObject gets the k8s resource o from a provider kubernetes object kon (Kube Object Name)
// from the desired array of the FunctionIO.
func (d *DesiredResources[T, O]) GetFromKubeObject(ctx context.Context, o client.Object, kon string) error {
	ko, err := getKubeObjectFrom(ctx, &d.Resources, kon)
	if err != nil {
		return err
	}
	return fromKubeObject(ko, o)
}

// ResourceExists check weather a relevant resource exists in this slice.
// A relevant resource is any resource that is not a Kubernetes Object resource.
// The function also checks resources inside Kubernetes Objects in case unmarshalling
// does not fail.
func (d *DesiredResources[T, O]) ResourceExists(name string) bool {
	for _, r := range d.Resources {
		var o client.Object
		err := json.Unmarshal(r.GetRaw(), o)
		if err != nil {
			return false
		}
		if o.GetObjectKind().GroupVersionKind() == xkube.ObjectGroupVersionKind {
			ko := o.(*xkube.Object)
			var o client.Object
			err = json.Unmarshal(ko.Spec.ForProvider.Manifest.Raw, o)
			if err != nil {
				return false
			}
			if o.GetName() == name {
				return true
			}
		} else {
			if r.GetName() == name {
				return true
			}
		}
	}
	return false
}

// PutIntoKubeObject adds or updates the desired resource into its kube object
func (d *DesiredResources[T, O]) PutIntoKubeObject(ctx context.Context, o client.Object, kon string, refs ...xkube.Reference) error {
	log := controllerruntime.LoggerFrom(ctx)

	ko := &xkube.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       xkube.ObjectKind,
			APIVersion: xkube.ObjectKindAPIVersion,
		},
		Spec: xkube.ObjectSpec{
			References: refs,
		},
	}
	err := getFrom(ctx, &d.Resources, ko, kon)
	if err != nil && err != ErrNotFound {
		return err
	}

	log.V(1).Info("Put object into kube object", "object", o, "kube object name", kon)
	err = updateKubeObject(o, ko)
	if err != nil {
		return err
	}

	return d.put(ko, kon)
}

// GetManagedResource will unmarshall the resource from the desired array.
// This will return any changes that a previous function has made to the desired array.
func (d *DesiredResources[T, O]) GetManagedResource(ctx context.Context, resName string, obj client.Object) error {
	return getFrom(ctx, &d.Resources, obj, resName)
}

// PutManagedResource will add the object as is to the FunctionIO desired array.
// It assumes that the given object is adheres to Crossplane's ManagedResource model.
func (d *DesiredResources[T, O]) PutManagedResource(obj client.Object) error {
	return d.put(obj, obj.GetName())
}

func (d *DesiredResources[T, O]) put(obj client.Object, resName string) error {
	kind, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}

	obj.GetObjectKind().SetGroupVersionKind(kind[0])
	rawData, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	for _, res := range d.Resources {
		if res.GetName() == resName {
			res.SetRaw(rawData)
			return nil
		}
	}

	d.Resources = append(d.Resources, desiredResource(
		xfnv1alpha1.DesiredResource{
			Name: resName,
			Resource: runtime.RawExtension{
				Raw: rawData,
			},
		},
	))
	return nil
}

// desiredResource is a wrapper around xfnv1alpha1.DesiredResource
// so we can satisfy the Resource interface.
type desiredResource xfnv1alpha1.DesiredResource

func (d desiredResource) GetName() string {
	return d.Name
}

func (d desiredResource) GetRaw() []byte {
	return d.Resource.Raw
}

func (d desiredResource) SetRaw(raw []byte) {
	d.Resource.Raw = raw
}
