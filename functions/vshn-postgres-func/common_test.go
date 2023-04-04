package vshnpostgres

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/vshn/appcat-comp-functions/runtime"
	"sigs.k8s.io/yaml"
)

func getFunctionFromFile(t assert.TestingT, file string) *runtime.Runtime {
	p, _ := filepath.Abs(".")
	before, _, _ := strings.Cut(p, "/functions")
	dat, err := os.ReadFile(before + "/test/transforms/vshn-postgres/" + file)
	assert.NoError(t, err)

	funcIO := runtime.Runtime{}
	err = yaml.Unmarshal(dat, &funcIO)
	assert.NoError(t, err)

	return &funcIO
}

func getCompositeFromIO[T any](t assert.TestingT, io *runtime.Runtime, obj *T) *T {
	err := json.Unmarshal(io.Observed.Composite.Resource.Raw, obj)
	assert.NoError(t, err)

	return obj
}
