//go:build !dolt_embedded

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

// connectEmbedded returns an error when built without dolt_embedded tag.
func connectEmbedded(_ string) (*doltstore.DoltStore, error) {
	return nil, fmt.Errorf("embedded Dolt requires building with: go build -tags dolt_embedded")
}

// hasEmbeddedSupport returns false when built without dolt_embedded tag.
func hasEmbeddedSupport() bool {
	return false
}

func dbInitEmbedded(cmd *cobra.Command, _ string) error {
	return fmt.Errorf("embedded Dolt requires building with: go build -tags dolt_embedded")
}
