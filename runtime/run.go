package runtime

import (
	"context"
	"fmt"

	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/yaml"
)

// Exec reads FunctionIO from stdin and return the desired state via transform function
func Exec(ctx context.Context, log logr.Logger, runtime *Runtime, transform Transform) error {

	log.V(1).Info("Executing transformation function")
	res := transform.TransformFunc(ctx, runtime).Resolve()
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
func printFunctionIO(iof *xfnv1alpha1.FunctionIO, log logr.Logger) error {
	log.V(1).Info("Marshalling FunctionIO")
	fnc, err := yaml.Marshal(iof)
	if err != nil {
		return fmt.Errorf("failed to marshal function io: %w", err)
	}

	fmt.Println(string(fnc))
	return nil
}

func RunCommand(ctx *cli.Context, transforms []Transform) error {
	log := logr.FromContextOrDiscard(ctx.Context)

	log.V(1).Info("Creating new runtime")
	funcIO, err := NewRuntime(ctx.Context)
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
