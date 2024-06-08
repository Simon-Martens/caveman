package app

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/db/accesstokens"
	"github.com/Simon-Martens/caveman/db/datastore"
	"github.com/Simon-Martens/caveman/db/sessions"
	"github.com/Simon-Martens/caveman/db/users"
	"github.com/Simon-Martens/caveman/frontend"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/security"
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

	users     *users.UserManager
	sessions  *sessions.SessionManager
	datastore *datastore.DataStoreManager
	tokens    *accesstokens.AccessTokenManager

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
	if err := os.MkdirAll(a.dataDir, os.ModePerm); err != nil {
		return err
	}

	if err := a.InitDataDB(
		models.DEFAULT_DATA_DIR,
		models.DEFAULT_DATA_FILE,
		models.DEFAULT_DATA_MAX_OPEN_CONNS,
		models.DEFAULT_DATA_MAX_IDLE_CONNS,
	); err != nil {
		return err
	}

	// if err := a.initLogsDB(); err != nil {
	// 	return err
	// }

	if err := a.InitLogger(); err != nil {
		return err
	}

	if err := a.InitDataStoreManager(
		a.db,
		models.DEFAULT_DATASTORE_TABLE,
		models.DEFAULT_ID_FIELD,
	); err != nil {
		return err
	}

	if err := a.InitSettings(a.datastore, models.DATASTORE_SETTINGS_KEY); err != nil {
		return err
	}

	if err := a.InitUsers(
		a.db,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_USER_EXPIRATION,
		a.settings,
	); err != nil {
		return err
	}

	if err := a.InitSessions(
		a.db,
		models.DEFAULT_SESSIONS_TABLE,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_SESSION_EXPIRATION,
		a.settings,
	); err != nil {
		return err
	}

	if err := a.InitTokens(
		a.db,
		models.DEFAULT_ACCESS_TOKENS_TABLE,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_RESOURCE_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_RESOURCE_SESSION_EXPIRATION,
	); err != nil {
		return err
	}

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

func (app *App) Sessions() *sessions.SessionManager {
	return app.sessions
}

func (app *App) Users() *users.UserManager {
	return app.users
}

func (app *App) DataDir() string {
	return app.dataDir
}

func (app *App) DB() *db.DB {
	return app.db
}

func (a *App) Settings() *models.Settings {
	return a.settings
}

func (a *App) Registry() *templates.Registry {
	return a.registry
}

// INFO: every init function must make sure of it's own dependencies
func (app *App) InitLogger() error {
	// TODO: create a handler, to write logs to the db
	app.logger = slog.Default()
	return nil
}

func (app *App) InitDataDB(dir string, file string, maxopenconns int, maxidleconns int) error {
	p := filepath.Join(dir, file)
	db, err := db.New(p, maxopenconns, maxidleconns)
	if err != nil {
		return err
	}

	app.db = db

	if app.isDev {
		db.ConnectLogger()
	}

	return nil
}

func (a *App) InitDataStoreManager(db *db.DB, tablename string, idfield string) error {
	ds, err := datastore.New(db, tablename, idfield)
	if err != nil {
		return err
	}
	a.datastore = ds
	return nil
}

func (a *App) InitSettings(dsm *datastore.DataStoreManager, key string) error {
	if dsm == nil {
		return errors.New("datastore manager is nil")
	}

	// TODO: its a bit bad that key is hardcoded here and migt not match settings.Key()
	s, err := dsm.SelectLatest(key)
	if err == datastore.ErrNotFound {
		sets := models.DefaultSettings()
		s, err = dsm.Insert(sets)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	sets := &models.Settings{}
	if err := json.Unmarshal(s.Data, sets); err != nil {
		return err
	}

	a.settings = sets
	return nil
}

func (a *App) InitSessions(db *db.DB, stn, utn, idfield string, lsessexp, ssessexp int, sets *models.Settings) error {
	if sets == nil || db == nil {
		return errors.New("settings or db is nil")
	}
	if sets.SessionSeed == 0 {
		sets.SessionSeed = security.GenRandomUIntNotPrime()
	}
	sm, err := sessions.New(db, stn, utn, idfield, lsessexp, ssessexp, sets.SessionSeed)
	if err != nil {
		return err
	}
	a.sessions = sm
	return nil
}

func (a *App) InitUsers(db *db.DB, utn, idfield string, uexp int, sets *models.Settings) error {
	if sets == nil || db == nil {
		return errors.New("settings or db is nil")
	}
	if sets.UserSeed == 0 {
		sets.UserSeed = security.GenRandomUIntNotPrime()
	}
	um, err := users.New(db, utn, idfield, uexp, sets.UserSeed)
	if err != nil {
		return err
	}
	a.users = um
	return nil
}

func (a *App) InitTokens(db *db.DB, atn, utn, idfield string, lressexp, sressexp int) error {
	tm, err := accesstokens.New(db, atn, utn, idfield, lressexp, sressexp)
	if err != nil {
		return err
	}
	a.tokens = tm
	return nil
}
