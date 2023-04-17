package vshnpostgres

import (
	"context"
	xfnv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"github.com/vshn/appcat-comp-functions/runtime"
	v1 "github.com/vshn/component-appcat/apis/vshn/v1"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func loadRuntimeFromFile(t assert.TestingT, file string) *runtime.Runtime[v1.VSHNPostgreSQL, *v1.VSHNPostgreSQL] {
	p, _ := filepath.Abs(".")
	before, _, _ := strings.Cut(p, "/functions")
	f, err := os.Open(before + "/test/transforms/vshn-postgres/" + file)
	assert.NoError(t, err)
	os.Stdin = f
	funcIO, err := runtime.NewRuntime[v1.VSHNPostgreSQL, *v1.VSHNPostgreSQL](context.Background())
	assert.NoError(t, err)

	return funcIO
}

func getFunctionIo(funcIO *runtime.Runtime[v1.VSHNPostgreSQL, *v1.VSHNPostgreSQL]) xfnv1alpha1.FunctionIO {
	field := reflect.ValueOf(funcIO).Elem().FieldByName("io")
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(xfnv1alpha1.FunctionIO)
}
