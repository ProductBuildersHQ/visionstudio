package evidence

import (
	"testing"
)

func TestParseTrailer(t *testing.T) {
	tests := []struct {
		value string
		want  []string
	}{
		{"RMI-PRISMCONTROL-005", []string{"RMI-PRISMCONTROL-005"}},
		{"RMI-A-001, RMI-B-002", []string{"RMI-A-001", "RMI-B-002"}},
		{"", nil},
		{"no refs here", nil},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got := ParseTrailer(tt.value)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseTrailer(%q) = %v, want %v", tt.value, got, tt.want)
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Fatalf("ParseTrailer(%q)[%d] = %q, want %q", tt.value, i, g, tt.want[i])
				}
			}
		})
	}
}

func TestParseCommitMessage(t *testing.T) {
	msg := `feat(compositor): add alpha-channel overlay support

Implements transparent WebM input handling for avatar compositing.

Refs: RMI-VIDEOASCODE-019`

	refs := ParseCommitMessage(msg)
	if len(refs) != 1 || refs[0] != "RMI-VIDEOASCODE-019" {
		t.Fatalf("expected [RMI-VIDEOASCODE-019], got %v", refs)
	}
}

func TestParseCommitMessageMultipleRefs(t *testing.T) {
	msg := `fix: resolve compatibility issue

Refs: RMI-A-001, RMI-B-002`

	refs := ParseCommitMessage(msg)
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %v", refs)
	}
}

func TestParseCommitMessageNoTrailer(t *testing.T) {
	msg := `chore: update dependencies`
	refs := ParseCommitMessage(msg)
	if len(refs) != 0 {
		t.Fatalf("expected no refs, got %v", refs)
	}
}

func TestParseBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"rmi/prismcontrol-013-mcp-server", "RMI-PRISMCONTROL-013"},
		{"rmi/omnidevxcore-003", "RMI-OMNIDEVXCORE-003"},
		{"rmi/A-001", "RMI-A-001"},
		{"main", ""},
		{"feature/add-widget", ""},
		{"rmi/", ""},
		{"rmi/bad", ""},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := ParseBranch(tt.branch)
			if got != tt.want {
				t.Fatalf("ParseBranch(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestAttribute(t *testing.T) {
	tests := []struct {
		name    string
		trailer string
		branch  string
		wantIDs []string
		wantSrc string
	}{
		{"trailer wins", "RMI-A-001", "rmi/b-002-slug", []string{"RMI-A-001"}, "trailer"},
		{"branch fallback", "", "rmi/prismcontrol-013-mcp", []string{"RMI-PRISMCONTROL-013"}, "branch"},
		{"no match", "", "main", nil, ""},
		{"trailer multi", "RMI-A-001, RMI-B-002", "", []string{"RMI-A-001", "RMI-B-002"}, "trailer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Attribute(tt.trailer, tt.branch)
			if got.Source != tt.wantSrc {
				t.Fatalf("source = %q, want %q", got.Source, tt.wantSrc)
			}
			if len(got.RMIIDs) != len(tt.wantIDs) {
				t.Fatalf("RMIIDs = %v, want %v", got.RMIIDs, tt.wantIDs)
			}
			for i, id := range got.RMIIDs {
				if id != tt.wantIDs[i] {
					t.Fatalf("RMIIDs[%d] = %q, want %q", i, id, tt.wantIDs[i])
				}
			}
		})
	}
}
