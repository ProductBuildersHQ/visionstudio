//go:build dolt

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/visionstudio/pkg/service"
	"github.com/ProductBuildersHQ/visionstudio/pkg/store/doltstore"
)

func connectService(cmd *cobra.Command) (*service.Service, func(), error) {
	dataDir := getDataDir(cmd)
	if dataDir != "" {
		ds, err := doltstore.NewEmbedded(dataDir)
		if err != nil {
			return nil, nil, fmt.Errorf("connect (embedded): %w", err)
		}
		svc := service.New(ds)
		return svc, func() {
			if err := ds.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
			}
		}, nil
	}

	dsn := getDSN(cmd)
	ds, err := doltstore.New(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("connect (server): %w", err)
	}
	svc := service.New(ds)
	return svc, func() {
		if err := ds.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: close database: %v\n", err)
		}
	}, nil
}
