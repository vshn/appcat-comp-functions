package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObservedResources struct {
	resources []Resource
	composite xfnv1alpha1.ObservedComposite
}

// GetFromKubeObject gets the k8s resource o from a provider kubernetes object kon (Kube Object Name)
// from the observed array of the FunctionIO.
func (o *ObservedResources) GetFromKubeObject(ctx context.Context, obj client.Object, kon string) error {
	ko, err := getKubeObjectFrom(ctx, &o.resources, kon)
	if err != nil {
		return err
	}
	return o.fromKubeObject(ctx, ko, obj)
}

// GetManagedResource unmarshalls the managed resource with the given name into the given object.
// It reads from the Observed array.
func (o *ObservedResources) GetManagedResource(ctx context.Context, obj client.Object, resName string) error {
	return getFrom(ctx, &o.resources, obj, resName)
}

// GetComposite unmarshalls the observed composite from the function io to the given object.
func (o *ObservedResources) GetComposite(_ context.Context, obj client.Object) error {
	err := json.Unmarshal(o.composite.Resource.Raw, obj)
	if err != nil {
		return fmt.Errorf("cannot unmarshall observed composite: %v", err)
	}
	return nil
}

// GetCompositeConnectionDetails returns the connection details of the observed composite
func (o *ObservedResources) GetCompositeConnectionDetails(_ context.Context) *[]xfnv1alpha1.ExplicitConnectionDetail {
	return &o.composite.ConnectionDetails
}

// ListResources return the list of managed resources from observed object
func (o *ObservedResources) ListResources(_ context.Context) []Resource {
	return o.resources
}

// fromKubeObject checks into status field instead of spec. The spec may not show all the relevant data
// and the status field will not be changed with multiple transformation functions
func (o *ObservedResources) fromKubeObject(ctx context.Context, kobj *xkube.Object, obj client.Object) error {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("Unmarshalling resource from observed kube object", "kube object", kobj, reflect.TypeOf(obj).Kind())
	if kobj.Status.AtProvider.Manifest.Raw == nil {
		return ErrNotFound
	}
	return json.Unmarshal(kobj.Status.AtProvider.Manifest.Raw, obj)
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
