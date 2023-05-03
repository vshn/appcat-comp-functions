//go:build tools
// +build tools

// Package tools is a place to put any tooling dependencies as imports.
// Go modules will be forced to download and install them.
package tools

import (
	// Add any build-time dependencies here with blank imports like `_ "package"`
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
