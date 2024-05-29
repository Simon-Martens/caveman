package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

func HandleStaticFS(fileSystem fs.FS, indexFallback bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := c.PathParam("*")

		tmpPath, err := url.PathUnescape(p)
		if err != nil {
			return fmt.Errorf("failed to unescape path variable: %w", err)
		}
		p = tmpPath

		// fs.FS.Open() considers name with prefix `/` as invalid
		name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))

		fileErr := c.FileFS(name, fileSystem)

		if fileErr != nil && indexFallback && errors.Is(fileErr, echo.ErrNotFound) {
			return c.FileFS("index.html", fileSystem)
		}

		return fileErr
	}
}
