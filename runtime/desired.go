package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DesiredResources struct {
	resources []Resource
	composite xfnv1alpha1.DesiredComposite
}

// GetFromKubeObject gets the k8s resource o from a provider kubernetes object kon (Kube Object Name)
// from the desired array of the FunctionIO.
func (d *DesiredResources) GetFromKubeObject(ctx context.Context, o client.Object, kon string) error {
	ko, err := getKubeObjectFrom(ctx, &d.resources, kon)
	if err != nil {
		return fmt.Errorf("cannot get resource from kube object: %v", err)
	}
	return fromKubeObject(ctx, ko, o)
}

// PutIntoKubeObject adds or updates the desired resource into its kube object
func (d *DesiredResources) PutIntoKubeObject(ctx context.Context, o client.Object, kon string, refs ...xkube.Reference) error {
	log := controllerruntime.LoggerFrom(ctx)

	ko := &xkube.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       xkube.ObjectKind,
			APIVersion: xkube.ObjectKindAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: kon,
		},
		Spec: xkube.ObjectSpec{
			References: refs,
		},
	}
	err := getFrom(ctx, &d.resources, ko, kon)
	if err != nil && err != ErrNotFound {
		return err
	}

	log.V(1).Info("Preparing to put object into kube object", "object", o, "kube object name", kon)
	err = updateKubeObject(o, ko)
	if err != nil {
		return err
	}

	return d.put(ctx, ko, kon)
}

// GetManagedResource unmarshalls the resource from the desired array.
// This will return any changes that a previous function has made to the desired array.
func (d *DesiredResources) GetManagedResource(ctx context.Context, obj client.Object, resName string) error {
	return getFrom(ctx, &d.resources, obj, resName)
}

// PutManagedResource adds the object as is to the FunctionIO desired array.
// It assumes that the given object is adheres to Crossplane's ManagedResource model.
func (d *DesiredResources) PutManagedResource(ctx context.Context, obj client.Object) error {
	return d.put(ctx, obj, obj.GetName())
}

func (d *DesiredResources) put(ctx context.Context, obj client.Object, resName string) error {
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("Putting object into kube object", "object", obj, "kube object name", resName)
	kind, _, err := s.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("cannot get object kinds from %s: %v", obj.GetName(), err)
	}

	obj.GetObjectKind().SetGroupVersionKind(kind[0])
	rawData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("cannot marshall object %s: %v", obj.GetName(), err)
	}

	for _, res := range d.resources {
		if res.GetName() == resName {
			log.V(1).Info("Updating existing kube object with resource", "object", obj, "kube object name", resName)
			res.SetRaw(rawData)
			return nil
		}
	}

	log.V(1).Info("No kube object found, adding new one with resource", "object", obj, "kube object name", resName)
	d.resources = append(d.resources, desiredResource(
		xfnv1alpha1.DesiredResource{
			Name: resName,
			Resource: runtime.RawExtension{
				Raw: rawData,
			},
		},
	))
	return nil
}

// GetComposite unmarshalls the desired composite from the function io to the given object.
func (d *DesiredResources) GetComposite(_ context.Context, obj client.Object) error {
	err := json.Unmarshal(d.composite.Resource.Raw, obj)
	if err != nil {
		return fmt.Errorf("cannot unmarshall desired composite: %v", err)
	}
	return nil
}

// SetComposite sets a new desired composite to the function from the given object.
func (d *DesiredResources) SetComposite(_ context.Context, obj client.Object) error {
	raw, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("cannot marshal desired composite: %v", err)
	}
	d.composite.Resource.Raw = raw
	return nil
}

// GetCompositeConnectionDetails returns the connection details of the desired composite
func (d *DesiredResources) GetCompositeConnectionDetails(_ context.Context) []xfnv1alpha1.ExplicitConnectionDetail {
	return d.composite.ConnectionDetails
}

// PutCompositeConnectionDetail appends a connection detail to the connection details slice
// of this desired composite
func (d *DesiredResources) PutCompositeConnectionDetail(ctx context.Context, cd xfnv1alpha1.ExplicitConnectionDetail) {
	log := controllerruntime.LoggerFrom(ctx)
	for i, c := range d.composite.ConnectionDetails {
		if cd.Name == c.Name {
			log.V(1).Info("Updating existing desired composite connection detail", "cd", cd)
			d.composite.ConnectionDetails[i] = cd
			return
		}
	}
	log.V(1).Info("Adding desired composite connection detail", "cd", cd)
	d.composite.ConnectionDetails = append(d.composite.ConnectionDetails, cd)
}

// RemoveCompositeConnectionDetail removes a connection detail from the slice of connection details
// of this desired composite
func (d *DesiredResources) RemoveCompositeConnectionDetail(ctx context.Context, cd xfnv1alpha1.ExplicitConnectionDetail) error {
	log := controllerruntime.LoggerFrom(ctx)
	cds := d.composite.ConnectionDetails
	for i, c := range cds {
		if cd.Name == c.Name {
			log.V(1).Info("Removing connection detail from desired connection details slice", "cd", cd)
			d.composite.ConnectionDetails = append(cds[:i], cds[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// ListResources returns the list of managed resources from desired object
func (d *DesiredResources) ListResources(_ context.Context) []Resource {
	return d.resources
}

// RemoveResource removes a resource by name from the managed resources
// expect an error if resource not found
func (d *DesiredResources) RemoveResource(ctx context.Context, name string) error {
	log := controllerruntime.LoggerFrom(ctx)
	for i, r := range d.resources {
		if r.GetName() == name {
			log.V(1).Info("Removing resource from desired resources", "resource name", name)
			d.resources = append(d.resources[:i], d.resources[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
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
