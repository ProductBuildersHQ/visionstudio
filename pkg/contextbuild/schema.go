package contextbuild

import (
	_ "embed"
	"encoding/json"

	"github.com/invopop/jsonschema"
)

//go:embed contextpackage.schema.json
var contextPackageSchemaJSON []byte

//go:embed phasehandoff.schema.json
var phaseHandoffSchemaJSON []byte

// ContextPackageSchema returns the embedded JSON Schema for ContextPackage.
func ContextPackageSchema() []byte {
	return contextPackageSchemaJSON
}

// PhaseHandoffSchema returns the embedded JSON Schema for PhaseHandoffProjection.
func PhaseHandoffSchema() []byte {
	return phaseHandoffSchemaJSON
}

// GenerateSchema generates JSON Schema from the ContextPackage Go type.
// Run this to regenerate contextpackage.schema.json when types change:
//
//	go run -exec "sh -c" -<<'EOF'
//	package main
//	import (
//		"fmt"
//		"github.com/ProductBuildersHQ/visionstudio/pkg/contextbuild"
//	)
//	func main() {
//		out, _ := contextbuild.GenerateSchema()
//		fmt.Println(string(out))
//	}
//	EOF
func GenerateSchema() ([]byte, error) {
	r := jsonschema.Reflector{
		DoNotReference: false,
	}

	schema := r.Reflect(&ContextPackage{})
	schema.ID = "https://github.com/ProductBuildersHQ/prism-build/schema/contextpackage"
	schema.Title = "ContextPackage"
	schema.Description = "Deterministic context package for agent sessions, assembled from the PRISM execution graph and repository specs."

	return json.MarshalIndent(schema, "", "  ")
}
