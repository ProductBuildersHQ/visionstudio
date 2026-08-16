// Package specworkflow provides workflow loading from visionspec.
package specworkflow

import (
	"fmt"
	"sort"

	"github.com/ProductBuildersHQ/visionspec/pkg/template"
	"github.com/ProductBuildersHQ/visionspec/pkg/workflow"
	"github.com/ProductBuildersHQ/visionspec/pkg/workflows"
	"github.com/plexusone/structured-evaluation/rubric"
)

// WorkflowInfo summarizes a workflow for listing.
type WorkflowInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Extends       string   `json:"extends,omitempty"`
	SpecsRequired []string `json:"specs_required"`
	SpecsOptional []string `json:"specs_optional"`
}

// LoadedWorkflow wraps visionspec's LoadedWorkflow.
type LoadedWorkflow = workflows.LoadedWorkflow

// Loader wraps the workflow loader with optional custom directory.
type Loader struct {
	loader workflows.Loader
}

// NewLoader creates a loader that checks customDir first, then embedded defaults.
// If customDir is empty, only embedded workflows are available.
func NewLoader(customDir string) *Loader {
	var loader workflows.Loader
	if customDir != "" {
		loader = workflows.NewResolvingLoader(workflows.NewChainLoader(
			workflows.NewFileLoader(customDir),
			workflows.DefaultLoader(),
		))
	} else {
		loader = workflows.DefaultLoader()
	}
	return &Loader{loader: loader}
}

// DefaultLoader returns a loader using only embedded workflows.
func DefaultLoader() *Loader {
	return &Loader{loader: workflows.DefaultLoader()}
}

// Load returns a workflow by ID with inheritance resolved.
func (l *Loader) Load(id string) (*LoadedWorkflow, error) {
	return l.loader.Load(id)
}

// Available returns all available workflow IDs.
func (l *Loader) Available() []string {
	return l.loader.Available()
}

// List returns info for all available workflows. Spec lists are ordered by
// the workflow's execution sequence where one is defined (alphabetical
// otherwise) so output is deterministic.
func (l *Loader) List() ([]WorkflowInfo, error) {
	var result []WorkflowInfo
	for _, id := range l.loader.Available() {
		w, err := l.loader.Load(id)
		if err != nil {
			continue
		}
		var required, optional []string
		for specType, req := range w.Workflow.SpecConfig {
			if req.Required {
				required = append(required, specType)
			} else {
				optional = append(optional, specType)
			}
		}
		info := WorkflowInfo{
			ID:            id,
			Name:          w.Workflow.Name,
			Description:   w.Workflow.Description,
			Extends:       w.Workflow.Extends,
			SpecsRequired: orderBySequence(w.Workflow, required),
			SpecsOptional: orderBySequence(w.Workflow, optional),
		}
		result = append(result, info)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

// orderBySequence sorts spec-type IDs by the workflow's execution sequence,
// with types outside the sequence trailing in alphabetical order.
func orderBySequence(w *workflow.Workflow, types []string) []string {
	ordinal := map[string]int{}
	if w.Execution != nil {
		for i, specType := range w.Execution.Sequence {
			ordinal[specType] = i
		}
	}
	sort.SliceStable(types, func(i, j int) bool {
		oi, iok := ordinal[types[i]]
		oj, jok := ordinal[types[j]]
		switch {
		case iok && jok:
			return oi < oj
		case iok:
			return true
		case jok:
			return false
		default:
			return types[i] < types[j]
		}
	})
	return types
}

// GetWorkflow loads a workflow by ID.
func (l *Loader) GetWorkflow(id string) (*workflow.Workflow, error) {
	loaded, err := l.loader.Load(id)
	if err != nil {
		return nil, err
	}
	return loaded.Workflow, nil
}

// GetRubrics returns rubrics for a workflow, keyed by spec type.
func (l *Loader) GetRubrics(workflowID string) (map[string]*rubric.RubricSet, error) {
	loaded, err := l.loader.Load(workflowID)
	if err != nil {
		return nil, err
	}
	return loaded.Rubrics, nil
}

// GetRubric returns a specific rubric for a workflow and spec type.
func (l *Loader) GetRubric(workflowID, specType string) (*rubric.RubricSet, error) {
	rubrics, err := l.GetRubrics(workflowID)
	if err != nil {
		return nil, err
	}
	r, ok := rubrics[specType]
	if !ok {
		return nil, fmt.Errorf("no rubric for spec type %q in workflow %q", specType, workflowID)
	}
	return r, nil
}

// GetTemplates returns templates for a workflow, keyed by spec type.
func (l *Loader) GetTemplates(workflowID string) (map[string]*template.Template, error) {
	loaded, err := l.loader.Load(workflowID)
	if err != nil {
		return nil, err
	}
	return loaded.Templates, nil
}

// GetTemplate returns a specific template for a workflow and spec type.
func (l *Loader) GetTemplate(workflowID, specType string) (*template.Template, error) {
	templates, err := l.GetTemplates(workflowID)
	if err != nil {
		return nil, err
	}
	t, ok := templates[specType]
	if !ok {
		return nil, fmt.Errorf("no template for spec type %q in workflow %q", specType, workflowID)
	}
	return t, nil
}

// GetSynthesisGuidance returns the synthesis guidance for a spec type.
func (l *Loader) GetSynthesisGuidance(workflowID, specType string) (sources []string, guidance string, err error) {
	loaded, err := l.loader.Load(workflowID)
	if err != nil {
		return nil, "", err
	}
	if loaded.Workflow.Synthesis == nil {
		return nil, "", nil
	}
	rule, ok := loaded.Workflow.Synthesis[specType]
	if !ok {
		return nil, "", nil
	}
	// Profiles express LLM context as either `guidance` or `prompt_context`
	// (e.g. aws-two-way-door uses only the latter); honor both.
	g := rule.Guidance
	if g == "" {
		g = rule.PromptContext
	}
	return rule.Sources, g, nil
}
