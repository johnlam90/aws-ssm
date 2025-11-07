package main

import (
	"fmt"
	"os"

	"github.com/aws-ssm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if _, writeErr := fmt.Fprintf(os.Stderr, "Error: %v\n", err); writeErr != nil {
			// If we can't even print the error, just exit with error code
			os.Exit(1)
		}
		os.Exit(1)
	}
}
