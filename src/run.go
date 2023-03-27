package src

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Exec reads FunctionIO from stdin and return the desired state via transform function
func Exec[T any, O interface {
	client.Object
	*T
}](ctx context.Context, transform func(c context.Context, io *IO, obj O) (O, error), sb ...runtime.SchemeBuilder) error {
	log := controllerruntime.LoggerFrom(ctx)

	log.V(1).Info("Preparing to get FunctionIO from stdin")
	iof, err := NewFunctionIO(ctx)
	if err != nil {
		return fmt.Errorf("failed to get new IO: %w", err)
	}

	log.V(1).Info("Unmarshalling composite from FunctionIO")
	var t T
	obj := &t
	err = json.Unmarshal(iof.F.Observed.Composite.Resource.Raw, obj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal composite: %w", err)
	}

	log.Info("Executing transformation function")
	res, err := transform(ctx, iof, obj)
	if err != nil {
		return fmt.Errorf("failed to run transform function: %w", err)
	}

	log.V(1).Info("Marshalling composite")
	raw, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal composite: %w", err)
	}
	iof.F.Desired.Composite.Resource.Raw = raw

	log.V(1).Info("Marshalling FunctionIO")
	fnc, err := yaml.Marshal(iof.F)
	if err != nil {
		return fmt.Errorf("failed to marshal function io: %w", err)
	}

	log.V(1).Info("Output", "functionIO", string(fnc))
	return nil
}
