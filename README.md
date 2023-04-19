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
