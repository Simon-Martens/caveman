package app

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Simon-Martens/caveman/db"
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

	DEFAULT_DATA_MAX_OPEN_CONNS int = 120
	DEFAULT_DATA_MAX_IDLE_CONNS int = 20
	DEFAULT_LOGS_MAX_OPEN_CONNS int = 10
	DEFAULT_LOGS_MAX_IDLE_CONNS int = 2

	DEFAULT_LOCAL_STORAGE_DIR_NAME string = "storage"
	DEFAULT_BACKUPS_DIR_NAME       string = "backups"

	DEFAULT_DEV_MODE       bool   = false
	DEFAULT_DATA_DIR_NAME  string = "cm_data"
	DEFAULT_DATA_FILE_NAME string = "data.db"
)

type App struct {
	RootCmd *cobra.Command

	registry *templates.Registry
	store    *store.Store[any]
	settings *models.Settings
	logger   *slog.Logger
	db       *db.DB

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

func (a *App) Bootstrap() error {
	a.registry = templates.NewRegistry(frontend.RoutesFS)
	a.store = store.New(map[string]interface{}{})

	// clear resources of previous core state (if any)
	if err := a.ResetBootstrapState(); err != nil {
		return err
	}

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
	// if err := a.initLogger(); err != nil {
	// 	return err
	// }

	_ = a.RefreshSetupState()
	_, _ = a.RefreshSettings()
	return nil
}

func (a *App) Terminate() error {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
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

func (a *App) IsBootstrapped() bool {
	return a.store != nil
}

func (a *App) ResetBootstrapState() error {
	// TODO:
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
	// maxOpenConns := DEFAULT_DATA_MAX_OPEN_CONNS
	// maxIdleConns := DEFAULT_DATA_MAX_IDLE_CONNS

	db, err := db.New(filepath.Join(app.dataDir, DEFAULT_DATA_FILE_NAME))
	if err != nil {
		return err
	}

	app.db = db

	// if app.IsDev() {
	// 	nonconcurrentDB.QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
	// 		color.HiBlack("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
	// 	}
	// 	nonconcurrentDB.ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
	// 		color.HiBlack("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
	// 	}
	// 	concurrentDB.QueryLogFunc = nonconcurrentDB.QueryLogFunc
	// 	concurrentDB.ExecLogFunc = nonconcurrentDB.ExecLogFunc
	// }
	//
	// app.dao = app.createDaoWithHooks(concurrentDB, nonconcurrentDB)
	//
	// return nil
	return nil
}
