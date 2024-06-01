package app

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/db/sessions"
	"github.com/Simon-Martens/caveman/db/users"
	"github.com/Simon-Martens/caveman/frontend"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/store"
	"github.com/Simon-Martens/caveman/tools/templates"

	"github.com/spf13/cobra"
)

type App struct {
	RootCmd *cobra.Command

	registry *templates.Registry
	store    *store.Store[any]
	settings *models.Settings
	logger   *slog.Logger
	db       *db.DB

	users    *users.UserManager
	sessions *sessions.SessionManager

	isDev   bool
	dataDir string
}

func New(sets models.Config) *App {

	app := &App{
		dataDir: sets.DataDir,
		isDev:   sets.Dev,
	}
	return app
}

// Bootstrap initializes the application, including
// - database
// - store (a in memory key-value store)
// - logger
// - settings
// - built-in template registry
// We do not load user defined templates here; also no routes or middleware
// as the server is seperately initialized.
func (a *App) Bootstrap() error {

	// clear resources of previous core state (if any)
	if err := a.ResetBootstrapState(); err != nil {
		return err
	}

	a.registry = templates.NewRegistry(frontend.RoutesFS)
	a.store = store.New(map[string]interface{}{})

	// ensure that data dir exist
	log.Println("Data dir: ", a.dataDir)
	if err := os.MkdirAll(a.dataDir, os.ModePerm); err != nil {
		return err
	}

	if err := a.initDataDB(); err != nil {
		return err
	}

	// if err := a.initLogsDB(); err != nil {
	// 	return err
	// }
	//
	if err := a.initLogger(); err != nil {
		return err
	}

	_ = a.RefreshSetupState()
	_, _ = a.RefreshSettings()

	um, err := users.New(a.db, models.DEFAULT_USERS_TABLE_NAME, models.DEFAULT_ID_FIELD)
	if err != nil {
		return err
	}
	a.users = um

	sm, err := sessions.New(
		a.db,
		models.DEFAULT_SESSIONS_TABLE_NAME,
		models.DEFAULT_USERS_TABLE_NAME,
		models.DEFAULT_ID_FIELD)
	if err != nil {
		return err
	}
	a.sessions = sm

	return nil
}

func (a *App) Terminate() error {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
	}

	if a.logger != nil {

	}

	return nil
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
	s := a.store.Get(models.STORE_KEY_SETUP_STATE)
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

func (a *App) IsBootstrapped() bool {
	return a.store != nil || a.db != nil || a.logger != nil || a.registry != nil
}

func (a *App) ResetBootstrapState() error {
	a.store = nil
	a.logger = nil
	a.registry = nil

	// We do this last since it can err
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
	}

	return nil
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

// Logger returns the default app logger.
//
// If the application is not bootstrapped yet, fallbacks to slog.Default().
func (app *App) Logger() *slog.Logger {
	if app.logger == nil {
		return slog.Default()
	}

	return app.logger
}

// IsDev returns true if the application is running in development mode.
func (app *App) IsDev() bool {
	return app.isDev
}

func (app *App) initDataDB() error {
	p := filepath.Join(app.dataDir, models.DEFAULT_DATA_FILE_NAME)
	db, err := db.New(p, models.DEFAULT_DATA_MAX_OPEN_CONNS, models.DEFAULT_DATA_MAX_IDLE_CONNS)
	if err != nil {
		return err
	}

	app.db = db

	if app.isDev {
		db.ConnectLogger()
	}

	return nil
}

func (app *App) Sessions() *sessions.SessionManager {
	return app.sessions
}

func (app *App) Users() *users.UserManager {
	return app.users
}

func (app *App) initLogger() error {
	// TODO: create a handler, to write logs to the db

	app.logger = slog.Default()

	return nil

}

func (app *App) DataDir() string {
	return app.dataDir
}

func (app *App) DB() *db.DB {
	return app.db
}
