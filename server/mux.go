package server

import (
	"github.com/Simon-Martens/caveman/app"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewMux binds server wide middlewares; that wotk on all routes
func NewMux(app *app.App) *echo.Echo {
	mux := echo.New()

	// Set auth, role and HTMX headers
	// mux.Use(IntoRequestContext(app))

	mux.HTTPErrorHandler = ErrorHandler

	mux.Use(middleware.Logger())
	mux.Use(middleware.Recover())
	mux.Use(middleware.Secure())
	mux.Use(middleware.RemoveTrailingSlashWithConfig(
		middleware.TrailingSlashConfig{
			RedirectCode: 301,
		},
	))

	return mux
}
