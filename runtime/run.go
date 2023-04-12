package runtime

import (
	"context"
	"encoding/json"
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
}](ctx context.Context, log logr.Logger, iof *Runtime, transform func(c context.Context, log logr.Logger, io *Runtime, obj O) (O, error)) error {

	log.V(1).Info("Unmarshalling composite from FunctionIO")
	var t T
	obj := &t
	err := json.Unmarshal(iof.Func.Observed.Composite.Resource.Raw, obj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal composite: %w", err)
	}

	log.V(1).Info("Executing transformation function")
	res, err := transform(ctx, log, iof, obj)
	if err != nil {
		iof.AddResult(xfnv1alpha1.SeverityWarning, err.Error())
	}

	log.V(1).Info("Marshalling composite")
	raw, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal composite: %w", err)
	}
	iof.Func.Desired.Composite.Resource.Raw = raw

	return nil
}

// printFunctionIO prints the whole FunctionIO to stdout, so Crossplane can
// pick it up again.
func printFunctionIO(iof *Runtime, log logr.Logger) error {
	log.V(1).Info("Marshalling FunctionIO")
	fnc, err := yaml.Marshal(iof.Func)
	if err != nil {
		return fmt.Errorf("failed to marshal function io: %w", err)
	}

	fmt.Println(string(fnc))
	return nil
}

func RunCommand[T any, O interface {
	client.Object
	*T
}](ctx *cli.Context, transforms []Transform[T, O]) error {

	funcIO, err := setup(ctx)
	if err != nil {
		return err
	}

	log := logr.FromContextOrDiscard(ctx.Context)

	if ctx.String("function") != "" {
		for _, function := range transforms {
			if function.Name == ctx.String("function") {
				log.Info("Starting single function", "name", function.Name)
				err := Exec(ctx.Context, log, funcIO, function.TransformFunc)
				if err != nil {
					return err
				}
			}
		}
		return printFunctionIO(funcIO, log)
	}

	for _, function := range transforms {
		log.Info("Starting function", "name", function.Name)
		err := Exec(ctx.Context, log, funcIO, function.TransformFunc)
		if err != nil {
			return err
		}
	}

	return printFunctionIO(funcIO, log)
}

func setup(ctx *cli.Context) (*Runtime, error) {

	funcIO, err := NewRuntime(ctx.Context)
	if err != nil {
		return nil, err
	}

	return funcIO, nil
}

// NewFunctionFlag returns the "function" cli flag.
func NewFunctionFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "function",
		Usage:    "Name of the function to run. If not provided, all functions will run.",
		Required: false,
	}
}
