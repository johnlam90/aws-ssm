// Package main is the entry point for the aws-ssm CLI application.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/johnlam90/aws-ssm/cmd"
	"github.com/johnlam90/aws-ssm/pkg/metrics"
)

func main() {
	// Initialize global metrics service
	// This must be done explicitly here, not in init(), to avoid hidden background goroutines
	ctx := context.Background()
	metrics.InitializeGlobalMetricsService(ctx)

	if err := cmd.Execute(); err != nil {
		if _, writeErr := fmt.Fprintf(os.Stderr, "Error: %v\n", err); writeErr != nil {
			// If we can't even print the error, just exit with error code
			os.Exit(1)
		}
		os.Exit(1)
	}
}
