//go:build !dolt

package main

import (
	"context"

	"github.com/spf13/cobra"
)

// startPeriodicCommitter is a no-op when built without Dolt support.
func startPeriodicCommitter(_ context.Context, _ *cobra.Command) func() {
	return func() {}
}
