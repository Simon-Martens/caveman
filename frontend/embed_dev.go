//go:build dev

package frontend

import (
	"os"
)

const (
	STATIC_FILEPATH = "./frontend/_builtin"
	ROUTES_FILEPATH = "./frontend/_routes"
)

var StaticFS = os.DirFS(STATIC_FILEPATH)
var RoutesFS = os.DirFS(ROUTES_FILEPATH)
