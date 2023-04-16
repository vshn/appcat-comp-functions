package runtime

import (
	"context"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObservedResources[T any, O interface {
	client.Object
	*T
}] struct {
	Resources         []Resource
	Composite         T
	ConnectionDetails []xfnv1alpha1.ExplicitConnectionDetail
}

// GetFromKubeObject gets the k8s resource o from a provider kubernetes object kon (Kube Object Name)
// from the observed array of the FunctionIO.
func (o *ObservedResources[T, O]) GetFromKubeObject(ctx context.Context, obj client.Object, kon string) error {
	ko, err := getKubeObjectFrom(ctx, &o.Resources, obj, kon)
	if err != nil {
		return err
	}
	return fromKubeObject(ko, obj)
}

// GetManagedResource will unmarshall the managed resource with the given name into the given object.
// It reads from the Observed array.
func (o *ObservedResources[T, O]) GetManagedResource(resName string, obj client.Object) error {
	return getFrom(&o.Resources, obj, resName)
}

// observedResource is a wrapper around xfnv1alpha1.ObservedResource
// so we can satisfy the Resource interface.
type observedResource xfnv1alpha1.ObservedResource

func (o observedResource) GetName() string {
	return o.Name
}

func (o observedResource) GetRaw() []byte {
	return o.Resource.Raw
}

func (o observedResource) SetRaw(raw []byte) {
	o.Resource.Raw = raw
}

func (o observedResource) GetKind() schema.ObjectKind {
	return o.Resource.Object.GetObjectKind()
}
