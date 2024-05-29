package helpers

import (
	"github.com/Simon-Martens/caveman/app"
	"github.com/labstack/echo"
)

type DataFunc = func(app *app.App, c echo.Context) (map[string]any, error)

func BaseState(app *app.App, c echo.Context) (map[string]any, error) {
	s := app.Settings()
	setup := app.SetupState()
	if s == nil {
		return map[string]any{
			"Name":    "",
			"Desc":    "",
			"URL":     "",
			"Edition": "",
			"Contact": "",
			"Curie":   "",
			"Icon":    "",
			"Setup":   setup,
		}, nil
	}

	return map[string]any{
		"Name":    s.Name,
		"Desc":    s.Desc,
		"URL":     s.URL,
		"Edition": s.Edition,
		"Contact": s.Contact,
		"Curie":   s.Curie,
		"Icon":    s.Icon,
		"Setup":   setup,
	}, nil
}
