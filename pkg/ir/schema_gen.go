//go:build ignore

// Schema generator for VisionStudio IR types.
// Run with: go run schema_gen.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"

	"github.com/ProductBuildersHQ/visionstudio/pkg/ir"
)

func main() {
	schemaDir := "schema"
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", schemaDir, err)
		os.Exit(1)
	}

	schemas := []struct {
		name string
		typ  any
	}{
		{"repo-snapshot", ir.RepoSnapshot{}},
		{"execution-ir", ir.ExecutionIR{}},
		{"maturity-ir", ir.MaturityIR{}},
		{"roadmap-ir", ir.RoadmapIR{}},
		{"devx-ir", ir.DevXIR{}},
		{"contrib-ir", ir.ContribIR{}},
	}

	r := jsonschema.Reflector{
		DoNotReference: true,
	}

	for _, s := range schemas {
		schema := r.Reflect(s.typ)
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal %s: %v\n", s.name, err)
			os.Exit(1)
		}

		path := filepath.Join(schemaDir, s.name+".schema.json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("Generated %s\n", path)
	}
}
