//go:build !dev
// +build !dev

// Package ui handles the caveman Admin frontend embedding.
// we could use io/fs.Sub to get a sub filesystem, but it errors. echo.MustSubFS throws on error
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:assets
var ui_static embed.FS
var StaticFS = MustSubFS(ui_static, "assets")

//go:embed all:routes
var ui_routes embed.FS
var RoutesFS = MustSubFS(ui_routes, "routes")

func MustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)

	if err != nil {
		panic("Could not create SubFS for " + dir)
	}

	return sub
}
