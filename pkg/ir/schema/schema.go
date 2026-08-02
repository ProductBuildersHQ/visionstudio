// Package schema provides embedded JSON Schema definitions for VisionStudio IR types.
package schema

import _ "embed"

//go:embed repo-snapshot.schema.json
var RepoSnapshot []byte

//go:embed execution-ir.schema.json
var ExecutionIR []byte

//go:embed maturity-ir.schema.json
var MaturityIR []byte

//go:embed roadmap-ir.schema.json
var RoadmapIR []byte

//go:embed devx-ir.schema.json
var DevXIR []byte

//go:embed contrib-ir.schema.json
var ContribIR []byte
