// Package evidence implements commit trailer parsing, branch name
// extraction, and delivery evidence types for RMI attribution.
//
// Attribution precedence (TRD §8): trailer → branch name.
// PR body references are a future extension (GitHub API dependent).
package evidence

import (
	"regexp"
	"strings"
)

// TrailerKey is the git trailer key used for RMI attribution.
const TrailerKey = "Refs"

var (
	rmiRefPattern    = regexp.MustCompile(`RMI-[A-Z0-9]+-\d{3,}`)
	branchRMIPattern = regexp.MustCompile(`^rmi/([A-Za-z0-9]+-\d{3,})`)
)

// ParseTrailer extracts RMI IDs from a Refs: trailer value.
// The input is the trailer value only (not the key), e.g., "RMI-A-001, RMI-B-002".
func ParseTrailer(value string) []string {
	matches := rmiRefPattern.FindAllString(value, -1)
	return matches
}

// ParseCommitMessage extracts RMI references from a full commit message
// by scanning for Refs: trailers in the footer section (after the last
// blank line).
func ParseCommitMessage(msg string) []string {
	lines := strings.Split(msg, "\n")
	var refs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, TrailerKey+":"); found {
			refs = append(refs, ParseTrailer(after)...)
		}
	}
	return refs
}

// ParseBranch extracts an RMI ID from a branch name following the
// convention rmi/<repo>-<nnn>[-slug]. Returns "" if the branch
// doesn't match.
func ParseBranch(branch string) string {
	m := branchRMIPattern.FindStringSubmatch(branch)
	if m == nil {
		return ""
	}
	return "RMI-" + strings.ToUpper(m[1])
}

// Attribution holds the result of attributing a commit to RMI(s).
type Attribution struct {
	RMIIDs []string
	Source string // "trailer", "branch", or ""
}

// Attribute resolves RMI attribution for a commit using the precedence
// order: trailer → branch name. Returns an empty Attribution if no
// RMI reference is found.
func Attribute(trailerValue, branch string) Attribution {
	if refs := ParseTrailer(trailerValue); len(refs) > 0 {
		return Attribution{RMIIDs: refs, Source: "trailer"}
	}
	if rmi := ParseBranch(branch); rmi != "" {
		return Attribution{RMIIDs: []string{rmi}, Source: "branch"}
	}
	return Attribution{}
}
