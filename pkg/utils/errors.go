// Package utils provides shared utilities for azfloci.
package utils

import "fmt"

// CLIError wraps an Azure CLI execution error with exit code and stderr.
type CLIError struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("az command failed (exit %d): %s\nstderr: %s", e.ExitCode, e.Command, e.Stderr)
}

// AuthError indicates the user is not authenticated with Azure CLI.
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "not authenticated: run 'az login' first"
}

// ValidationError indicates invalid input or configuration.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// WorkflowError wraps errors during workflow execution.
type WorkflowError struct {
	StepName string
	Cause    error
}

func (e *WorkflowError) Error() string {
	return fmt.Sprintf("workflow step %q failed: %v", e.StepName, e.Cause)
}

func (e *WorkflowError) Unwrap() error {
	return e.Cause
}
