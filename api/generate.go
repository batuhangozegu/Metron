// Package api holds the code generation entry point for the Go service.
//
// buf lives in its own tools module so its dependencies do not change the
// versions this module builds with.
package api

//go:generate go tool -modfile=tools/go.mod buf generate
