package controllers

import "context"

// Interface represents a runnable component.
type Interface interface {
	// Run runs the component.
	Run(ctx context.Context, workers int) error
}
