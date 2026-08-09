// Package web embeds the built VisionStudio SPA (web/dist) so the visionstudio
// binary can serve the UI with no external files. Release builds run
// `npm run build` in web/ before compiling; a plain checkout embeds only the
// web/dist/.gitkeep placeholder, in which case DistFS reports the UI as not
// built and the server falls back to --web-dist / a disk web/dist, or prints a
// clear "build the UI" message.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// DistFS returns the embedded web/dist rooted at dist/, and true when a real
// SPA (index.html) was embedded. It returns false when only the placeholder is
// present — i.e. the frontend was not built before the binary was compiled.
func DistFS() (fs.FS, bool) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return nil, false
	}
	return sub, true
}
