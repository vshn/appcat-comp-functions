package runtime

import (
	"context"
	"fmt"

	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Exec reads FunctionIO from stdin and return the desired state via transform function
func Exec[T any, O interface {
	client.Object
	*T
}](log logr.Logger, runtime *Runtime[T, O], transform Transform[T, O]) error {

	log.V(1).Info("Executing transformation function")
	res := transform.TransformFunc(log, runtime).Resolve()
	if res.Severity == xfnv1alpha1.SeverityNormal {
		res.Message = fmt.Sprintf("Function %s ran successfully", transform.Name)
	}
	runtime.io.Results = append(runtime.io.Results, res)

	runtime.io.Desired.Composite.Resource.Raw = runtime.Desired.composite.Resource.Raw
	runtime.io.Desired.Composite.ConnectionDetails = runtime.Desired.composite.ConnectionDetails

	runtime.io.Desired.Resources = make([]xfnv1alpha1.DesiredResource, len(runtime.Desired.resources))
	for i, r := range runtime.Desired.resources {
		runtime.io.Desired.Resources[i] = xfnv1alpha1.DesiredResource(r.(desiredResource))
	}

	return nil
}

// printFunctionIO prints the whole FunctionIO to stdout, so Crossplane can
// pick it up again.
func printFunctionIO(iof *xfnv1alpha1.FunctionIO, log logr.Logger) ([]byte, error) {
	log.V(1).Info("Marshalling FunctionIO")
	fnc, err := yaml.Marshal(iof)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal function io: %w", err)
	}

	return fnc, nil
}

func RunCommand[T any, O interface {
	client.Object
	*T
}](ctx context.Context, transforms []Transform[T, O], input []byte) ([]byte, error) {
	log := logr.FromContextOrDiscard(ctx)

	log.V(1).Info("Creating new runtime")
	funcIO, err := NewRuntime[T, O](ctx, input)
	if err != nil {
		return []byte{}, err
	}

	for _, function := range transforms {
		log.Info("Starting function", "name", function.Name)
		err = Exec(log, funcIO, function)
		if err != nil {
			return []byte{}, err
		}
	}

	return printFunctionIO(&funcIO.io, log)
}

// NewFunctionFlag returns the "function" cli flag.
func NewFunctionFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "function",
		Usage:    "Name of the function to run. If not provided, all functions will run.",
		Required: false,
	}
}
