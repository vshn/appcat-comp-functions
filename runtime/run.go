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
}](ctx context.Context, log logr.Logger, iof *Runtime[T, O], transform Transform[T, O]) error {

	log.V(1).Info("Executing transformation function")
	err := transform.TransformFunc(ctx, log, iof)
	if err != nil {
		iof.AddResult(xfnv1alpha1.SeverityWarning, err.Error())
	}

	log.V(1).Info("Marshalling observed composite")
	raw, err := json.Marshal(iof.Observed.Composite)
	if err != nil {
		return fmt.Errorf("failed to marshal Desired composite: %w", err)
	}
	iof.io.Observed.Composite.Resource.Raw = raw
	iof.io.Observed.Composite.ConnectionDetails = iof.Observed.ConnectionDetails

	log.V(1).Info("Marshalling desired composite")
	dRaw, err := json.Marshal(iof.Desired.Composite)
	if err != nil {
		return fmt.Errorf("failed to marshal desired composite: %w", err)
	}
	iof.io.Desired.Composite.Resource.Raw = dRaw
	iof.io.Desired.Composite.ConnectionDetails = iof.Desired.ConnectionDetails

	return nil
}

// printFunctionIO prints the whole FunctionIO to stdout, so Crossplane can
// pick it up again.
func printFunctionIO(iof *xfnv1alpha1.FunctionIO, log logr.Logger) error {
	log.V(1).Info("Marshalling FunctionIO")
	fnc, err := yaml.Marshal(iof)
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
	log := logr.FromContextOrDiscard(ctx.Context)

	log.V(1).Info("Creating new runtime")
	funcIO, err := NewRuntime[T, O](ctx.Context)
	if err != nil {
		return err
	}

	if ctx.String("function") != "" {
		for _, function := range transforms {
			if function.Name == ctx.String("function") {
				log.Info("Starting single function", "name", function.Name)
				err = Exec(ctx.Context, log, funcIO, function)
				if err != nil {
					return err
				}
			}
		}
		return printFunctionIO(&funcIO.io, log)
	}

	for _, function := range transforms {
		log.Info("Starting function", "name", function.Name)
		err = Exec(ctx.Context, log, funcIO, function)
		if err != nil {
			return err
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
