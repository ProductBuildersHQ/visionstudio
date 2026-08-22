// Package version reports the running binary's version from Go's embedded
// build info (runtime/debug.ReadBuildInfo) rather than a hand-maintained
// constant. `go build`/`go install` stamp a module pseudo-version plus VCS
// revision/dirty-state automatically; this package just formats them.
// `go run` doesn't stamp VCS info, so String falls back to "dev" there.
package version

import "runtime/debug"

// String returns a human-readable version string, e.g.
// "v0.6.1-0.20260822171422-1aaf0f3 (1aaf0f3)" for a clean build --
// go's toolchain appends "+dirty" to the pseudo-version itself when the
// working tree had uncommitted changes at build time, so that signal
// doesn't need duplicating here. Falls back to "dev" when build info
// isn't available (go run, or a binary built with -buildvcs=false).
func String() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	v := info.Main.Version
	if v == "" || v == "(devel)" {
		v = "dev"
	}

	var revision string
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			revision = s.Value
			if len(revision) > 7 {
				revision = revision[:7]
			}
		}
	}

	if revision == "" {
		return v
	}
	return v + " (" + revision + ")"
}
