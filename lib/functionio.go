package lib

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Exec reads FunctionIO from stdin and return the desired state via transform function
func Exec[T any, O interface {
	client.Object
	*T
}](transform func(io *IO, o O) (O, error), sb ...runtime.SchemeBuilder) {
	iof, err := NewFunctionIO(sb...)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get new IO: %w", err))
	}

	var t T
	o := &t
	err = json.Unmarshal(iof.F.Observed.Composite.Resource.Raw, o)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get marshal object: %w", err))
	}

	res, err := transform(iof, o)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to run f: %w", err))
	}

	raw, err := json.Marshal(res)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get marshal object: %w", err))
	}
	iof.F.Desired.Composite.Resource.Raw = raw

	fnc, err := yaml.Marshal(iof.F)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get marshal function: %w", err))
	}

	_, _ = fmt.Println(string(fnc))
}
