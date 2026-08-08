//go:build ignore

// Schema generator for VisionStudio API types.
// Generates JSON Schema from Go types in pkg/apitypes.
//
// Usage: go generate ./pkg/apitypes
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/invopop/jsonschema"

	"github.com/ProductBuildersHQ/visionstudio/pkg/apitypes"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	r := jsonschema.Reflector{
		DoNotReference:             false,
		ExpandedStruct:             false,
		RequiredFromJSONSchemaTags: true,
	}

	// JudgeResult.Report uses rubric.Rubric directly from structured-evaluation.
	// The schema generator will include rubric types via the $ref in JudgeResult.
	types := []struct {
		name string
		typ  any
	}{
		{"ExecutionResponse", apitypes.ExecutionResponse{}},
		{"SpecsResponse", apitypes.SpecsResponse{}},
		{"SpecFilesResponse", apitypes.SpecFilesResponse{}},
		{"JudgeResult", apitypes.JudgeResult{}},
		{"Initiative", apitypes.Initiative{}},
		{"Program", apitypes.Program{}},
		{"Phase", apitypes.Phase{}},
		{"RoadmapItem", apitypes.RoadmapItem{}},
		{"ContextSpec", apitypes.ContextSpec{}},
		{"SpecWorkflow", apitypes.SpecWorkflow{}},
		{"SpecFile", apitypes.SpecFile{}},
		{"APIProgram", apitypes.APIProgram{}},
		{"APIInitiative", apitypes.APIInitiative{}},
		{"APIPhase", apitypes.APIPhase{}},
		{"APIRMI", apitypes.APIRMI{}},
		{"APIRepository", apitypes.APIRepository{}},
		{"APIRMIDependency", apitypes.APIRMIDependency{}},
		{"APIInitiativeDependency", apitypes.APIInitiativeDependency{}},
		{"APIStatusCount", apitypes.APIStatusCount{}},
	}

	outDir := "schema"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("create schema dir: %w", err)
	}

	defs := make(map[string]*jsonschema.Schema)

	for _, t := range types {
		schema := r.Reflect(t.typ)

		for name, def := range schema.Definitions {
			if existing, ok := defs[name]; ok {
				if existing.Type == "" && def.Type != "" {
					defs[name] = def
				}
			} else {
				defs[name] = def
			}
		}

		defs[t.name] = &jsonschema.Schema{
			Type:                 schema.Type,
			Properties:           schema.Properties,
			Required:             schema.Required,
			AdditionalProperties: schema.AdditionalProperties,
		}
	}

	combined := &jsonschema.Schema{
		Version:     "https://json-schema.org/draft/2020-12/schema",
		ID:          jsonschema.ID("https://visionstudio.productbuildershq.com/api.schema.json"),
		Definitions: defs,
	}

	data, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	outPath := filepath.Join(outDir, "api.schema.json")
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}

	fmt.Printf("Generated %s (%d definitions)\n", outPath, len(defs))

	for _, t := range types {
		schema := r.Reflect(t.typ)
		schema.Version = "https://json-schema.org/draft/2020-12/schema"
		schema.ID = jsonschema.ID(fmt.Sprintf("https://visionstudio.productbuildershq.com/%s.schema.json", t.name))

		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", t.name, err)
		}

		path := filepath.Join(outDir, t.name+".schema.json")
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", t.name, err)
		}
		fmt.Printf("Generated %s\n", path)
	}

	return nil
}

var _ = reflect.TypeOf
