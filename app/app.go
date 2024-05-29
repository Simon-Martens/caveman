package app

import (
	"github.com/Simon-Martens/caveman/frontend"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/store"
	"github.com/Simon-Martens/caveman/tools/templates"

	"github.com/spf13/cobra"
)

const (
	VERSION               = "0.1.0"
	STORE_KEY_SETUP_STATE = "setup"
	STATIC_FILEPATH       = "./frontend/assets"
	ROUTES_FILEPATH       = "./frontend/routes"
)

type App struct {
	RootCmd *cobra.Command

	registry *templates.Registry
	store    *store.Store[any]
	settings *models.Settings
}

func New(isDev bool, baseDir string) (*App, error) {

	app := &App{}
	return app.Bootstrap()
}

func (a *App) Bootstrap() (*App, error) {
	a.registry = templates.NewRegistry(frontend.RoutesFS)
	a.store = store.New(map[string]interface{}{})

	_ = a.RefreshSetupState()
	_, _ = a.RefreshSettings()
	return a, nil
}

func (a *App) RefreshSetupState() int {
	return 0
	// hasAdmins, err := db.HasAdmins(a.papp.Dao())
	//
	// if err != nil {
	// 	panic("Could not get the number of admins. " + err.Error())
	// }
	//
	// if !hasAdmins {
	// 	a.store.Set(STORE_KEY_SETUP_STATE, 1)
	// 	return 1
	// }
	//
	// hasSettings, err := db.HasSettings(a.papp.Dao())
	//
	// if err != nil {
	// 	panic("Could not get the number of settings. " + err.Error())
	// }
	//
	// if !hasSettings {
	// 	a.store.Set(STORE_KEY_SETUP_STATE, 2)
	// 	return 2
	// }
	//
	// a.store.Set(STORE_KEY_SETUP_STATE, 3)
	// return 3
}

func (a *App) SetupState() int {
	s := a.store.Get(STORE_KEY_SETUP_STATE)
	if s == nil {
		s = a.RefreshSetupState()
	}
	if casted, ok := s.(int); ok {
		return casted
	} else {
		panic("Unexpected type for setup state")
	}
}

func (a *App) Registry() *templates.Registry {
	return a.registry
}

func (a *App) Settings() *models.Settings {
	if a.settings == nil {
		a.RefreshSettings()
	}
	return a.settings
}

func (a *App) RefreshSettings() (*models.Settings, error) {
	// settings, err := db.LoadSettings(a.Dao())
	//
	// if err != nil {
	// 	return nil, err
	// }
	//
	// a.settings = settings

	return nil, nil
}
