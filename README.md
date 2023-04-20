# appcat-comp-functions

[![Build](https://img.shields.io/github/actions/workflow/status/vshn/appcat-comp-functions/.github/workflows/test.yml?branch=master)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/vshn/appcat-comp-functions)
[![Version](https://img.shields.io/github/v/release/vshn/appcat-comp-functions)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/vshn/appcat-comp-functions/total)][releases]

[build]: https://github.com/vshn/appcat-comp-functions/actions?query=workflow%3ATest
[releases]: https://github.com/vshn/appcat-comp-functions/releases
## Repository structure

```
.
├── docs
├── functions
│   ├── vshn-common-func
│   ├── vshn-postgres-func
│   └── vshn-redis-func
├── kind
├── runtime
└── test
```

- `./docs` contains relevant documentation in regard to this repository
- `./functions` contains the actual logic for each function-io. Each function-io can have multiple transformation go functions 
- `./runtime` contains a library with helper methods which helps with adding new functions. 
- `./kind` contains relevant files for local dev cluster
- `./test` contains test files

Check out the docs to understand how functions from this repository work.

## Add a new function-io

The framework is designed to easily add new composition functions to any AppCat service.
A function-io corresponds to one and only one composition thus multiple transformation go functions 
can be added to a function-io.
For instance, in `vshn-postgres-func` there are multiple transformation go functions such as `url` or `alerting`.


To add a new function to PostgreSQL by VSHN:

- Create a new package under `./functions/`.
- Create a go file and add a new transform go function to the list in `./cmd/<your-new-function-io>`.
- Implement the actual `Transform()` go function by using the helper functions from `runtime/desired.go` and `runtime/observed.go`.
- Register the transform go function in the `main.go`.
- Create a new app.go under `./cmd/<your-new-function-io>` and define a new `AppInfo` object.

This architecture allows us to run all the functions with a single command. But for debugging and development purpose it's possible to run each function separately, by using the `--function` flag.

## Manually testing a function
To test a function you can leverage the FunctionIO file in the `./test` folder.

`cat test/function-io.yaml | go run cmd/vshn-postgres-func/main.go --function myfunction > test.yaml`


## Usage of gRPC server - local development + kind cluster

entrypoint to start working with gRPC server is to run:
```
go run main.go -socket default.sock
```

it will create a socket file in Your local directory which is easier for development - no need to set permissions and directory structure.

It's also possible to trigger fake request to gRPC server by client (to imitate Crossplane):
```
cd test/grpc-client
go run main.go
```

if You want to run gRPC server in local kind cluster, please use:
1. [kindev](https://github.com/vshn/kindev). In makefile replace target:
   1. ```
      $(crossplane_sentinel): export KUBECONFIG = $(KIND_KUBECONFIG)
      $(crossplane_sentinel): kind-setup local-pv-setup
      # below line loads image to kind
	  kind load docker-image --name kindev ghcr.io/vshn/appcat-comp-functions
	  helm repo add crossplane https://charts.crossplane.io/stable
	  helm upgrade --install crossplane --create-namespace --namespace syn-crossplane crossplane/crossplane \
	  --set "args[0]='--debug'" \
	  --set "args[1]='--enable-composition-functions'" \
	  --set "args[2]='--enable-environment-configs'" \
	  --set "xfn.enabled=true" \
	  --set "xfn.args={--debug}" \
	  --set "xfn.image.repository=ghcr.io/vshn/appcat-comp-functions" \
	  --set "xfn.image.tag=latest" \
	  --wait
	  @touch $@   
      ```
2. [component-appcat](https://github.com/vshn/component-appcat) please append [file](https://github.com/vshn/component-appcat/blob/master/tests/golden/vshn/appcat/appcat/21_composition_vshn_postgres.yaml) with:
   1.   ```
        compositeTypeRef:
          apiVersion: vshn.appcat.vshn.io/v1
          kind: XVSHNPostgreSQL
        # we have to add functions declaration to postgresql
        functions:
          - container:
              image: postgresql
              runner:
                endpoint: unix-abstract:crossplane/fn/default.sock
            name: pgsql-func
            type: Container
        resources:
          - base:
            apiVersion: kubernetes.crossplane.io/v1alpha1
        ```

That's all - You can now run Your claims. This documentation and above workaround is just temporary solution, it should disappear once we actually implement composition functions. 