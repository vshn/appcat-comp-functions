package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	xkube "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type IO struct {
	F xfnv1alpha1.FunctionIO
	S *runtime.Scheme
}

func NewFunctionIO(builders ...runtime.SchemeBuilder) (*IO, error) {
	funcIO := xfnv1alpha1.FunctionIO{}

	s := runtime.NewScheme()

	for _, b := range builders {
		b.AddToScheme(s)
	}

	x, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(x, &funcIO)
	if err != nil {
		return nil, err
	}
	return &IO{
		F: funcIO,
		S: s,
	}, nil
}

func (in *IO) PutKubeObj(obj client.Object) error {
	o, err := in.kubeObject(obj)
	if err != nil {
		return err
	}
	return in.Put(o)
}

func (in *IO) Put(obj client.Object) error {
	name := obj.GetName()

	kind, _, err := in.S.ObjectKinds(obj)
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

var ErrNotFound = errors.New("not found")

func (in *IO) kubeObject(obj client.Object) (*xkube.Object, error) {
	kind, _, err := in.S.ObjectKinds(obj)
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
			Name: fmt.Sprintf("%s-%s", obj.GetNamespace(), obj.GetName()),
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

func (in *IO) FromKubeObject(kobj *xkube.Object, obj client.Object) error {
	if kobj.Spec.ForProvider.Manifest.Raw == nil {
		return nil
	}
	return json.Unmarshal(kobj.Spec.ForProvider.Manifest.Raw, obj)
}

func (in *IO) GetKubeObj(obj client.Object) error {
	ko := &xkube.Object{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", obj.GetNamespace(), obj.GetName()),
		},
	}
	err := in.Get(ko)
	if err != nil {
		return err
	}
	return err
}

func (in *IO) Get(obj client.Object) error {
	name := obj.GetName()

	// TODO: Some validation that the type matches

	for i, res := range in.F.Desired.Resources {
		if res.Name == name {
			return yaml.Unmarshal(in.F.Desired.Resources[i].Resource.Raw, obj)
		}
	}
	return ErrNotFound
}
