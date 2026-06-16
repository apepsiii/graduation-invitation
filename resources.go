package resources

import (
	"embed"
	"io/fs"
)

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:assets
var assetsFS embed.FS

var Version = "1.0.0"

func Templates() fs.FS {
	sub, _ := fs.Sub(templatesFS, "templates")
	return sub
}

func Assets() fs.FS {
	sub, _ := fs.Sub(assetsFS, "assets")
	return sub
}
