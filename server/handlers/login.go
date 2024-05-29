package handlers

import (
	"github.com/Simon-Martens/caveman/app"
	"github.com/Simon-Martens/caveman/server/helpers"
	"github.com/labstack/echo"
)

func HandleLogin(app *app.App) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Add("Clear-Site-Data", "\"*\"")
		return helpers.RenderTemplate(c, "login", helpers.BaseState, app)
	}
}
