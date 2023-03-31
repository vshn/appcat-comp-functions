# appcat-comp-functions

[![Build](https://img.shields.io/github/actions/workflow/status/vshn/appcat-comp-functions/.github/workflows/test.yml?branch=master)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/vshn/appcat-comp-functions)
[![Version](https://img.shields.io/github/v/release/vshn/appcat-comp-functions)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/vshn/appcat-comp-functions/total)][releases]

[build]: https://github.com/vshn/appcat-comp-functions/actions?query=workflow%3ATest
[releases]: https://github.com/vshn/appcat-comp-functions/releases
## Repository structure

This repository will build different docker images for different services. For that reason some folder structure is bound to the name of the service.

```
.
├── cmd
│   ├── vshn-postgres-func
│   └── vshn-redis-func
├── kind
├── functions
│   ├── vshn-common-func
│   ├── vshn-postgres-func
│   └── vshn-redis-func
├── runtime
└── test
```

- `./cmd` contains the entry point boilerplate for each service.
- `./pkg/functions` contains the actual logic for the function. Each transform should be in its own package.

## Add a new function

The framework is designed to easily add new composition functions to any AppCat service.

To add a new function to PostgreSQL by VSHN:

- Create a new package under `./pkg/functions/vshn-postgres-func`
- Add the transform function to the list in `./cmd/vshn-postgres-func`
- implement the actual `transform()` function by using the helper functions from `io.go`

This architecture allows us to run all the functions with a single command. But for debugging and development purpose it's possible to run each function seperately, by using the `--function` flag.

## Manually testing a function
To test a function you can leverage the FunctionIO file in the `./test` folder.

`cat test/function-io.yaml | go run cmd/vshn-postgres-func/main.go --function myfunction > test.yaml`
