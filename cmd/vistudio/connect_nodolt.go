//go:build !dolt

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
)

func connectService(_ *cobra.Command) (*service.Service, func(), error) {
	return nil, nil, fmt.Errorf("database support requires building with: go build -tags dolt")
}

func dbCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "db",
		Short: "Database management (requires -tags dolt build)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("database support requires building with: go build -tags dolt")
		},
	}
}
