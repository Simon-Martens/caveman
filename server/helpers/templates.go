package helpers

import (
	"net/http"

	"github.com/Simon-Martens/caveman/app"
	"github.com/labstack/echo/v4"
)

func RenderTemplate(c echo.Context, template string, data DataFunc, app *app.App) error {
	tmpl, err := app.Registry().LoadDir(template)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not find a valid template")
	}

	d, err := data(app, c)
	if err != nil {
		return err
	}

	html, err := tmpl.Render(d)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not find a valid template")

	}
	return c.HTML(http.StatusOK, html)
}

func RenderTemplatePartial(c echo.Context, template string, data DataFunc, app *app.App) error {
	tmpl, err := app.Registry().LoadFile(template)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	d, err := data(app, c)
	if err != nil {
		return err
	}

	html, err := tmpl.Render(d)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())

	}
	return c.HTML(http.StatusOK, html)
}
