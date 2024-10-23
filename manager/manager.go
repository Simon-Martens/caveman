package manager

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
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/security"
)

type Manager struct {
	cm_settings *models.Settings
	cm_db       *db.DB

	state *datastore.DataStoreManager

	DB       *db.DB
	logger   *slog.Logger
	users    *users.UserManager
	sessions *sessions.SessionManager
	tokens   *accesstokens.AccessTokenManager

	// These settings depend on startup settings, the settings above are read from the database
	isDev   bool
	dataDir string
}

func New(sets models.Config) *Manager {

	app := &Manager{
		dataDir: sets.DataDir,
		isDev:   sets.Dev,
	}
	return app
}

// Bootstrap initializes the application, including
// - database
// - logger
// - settings
// - built-in template registry
// We do not load user defined templates here; also no routes or middleware
// as the server is seperately initialized.
func (a *Manager) Bootstrap() error {

	// clear resources of previous core state (if any)
	if err := a.ResetBootstrapState(); err != nil {
		return err
	}

	// ensure that data dir exist
	if err := os.MkdirAll(a.dataDir, os.ModePerm); err != nil {
		return err
	}

	cm_db, err := a.CreateDB(
		a.dataDir,
		models.DEFAULT_DATA_FILE,
		models.DEFAULT_DATA_MAX_OPEN_CONNS,
		models.DEFAULT_DATA_MAX_IDLE_CONNS,
	)
	if err != nil {
		return err
	}

	a.cm_db = cm_db

	// if err := a.initLogsDB(); err != nil {
	// 	return err
	// }

	if err := a.InitLogger(); err != nil {
		return err
	}

	if err := a.BootstrapSettings(
		a.cm_db,
		models.DEFAULT_DATASTORE_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DATASTORE_SETTINGS_KEY,
	); err != nil {
		return err
	}

	if err := a.BootstrapAuth(
		a.cm_db,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ACCESS_TOKENS_TABLE,
		models.DEFAULT_SESSIONS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_SESSION_EXPIRATION,
		models.DEFAULT_LONG_RESOURCE_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_RESOURCE_SESSION_EXPIRATION,
		models.DEFAULT_USER_EXPIRATION,
	); err != nil {
		return err
	}

	return nil
}

func (a *Manager) BootstrapSettings(db *db.DB, tn, idf, key string) error {
	if err := a.InitDataStore(db, tn, idf); err != nil {
		return err
	}

	if err := a.InitSettings(a.state, key); err != nil {
		return err
	}

	return nil
}

func (a *Manager) BootstrapAuth(db *db.DB, tnu, tnat, tns, idf string, lseexp, sseexp, lrsexp, srsexp, uexp int) error {
	if err := a.InitUsers(db, tnu, idf, uexp, a.cm_settings); err != nil {
		return err
	}

	if err := a.InitTokens(db, tnat, tns, idf, lrsexp, srsexp); err != nil {
		return err
	}

	if err := a.InitSessions(db, tns, tnu, idf, lseexp, sseexp, a.cm_settings); err != nil {
		return err
	}

	return nil
}

func (a *Manager) Terminate() error {
	// 	if a.db != nil {
	// 		if err := a.db.Close(); err != nil {
	// 			return err
	// 		}
	// 	}
	//
	// 	if a.logger != nil {
	//
	// 	}
	//
	// 	return nil

	// TODO:
	return nil
}

func (a *Manager) RefreshSetupState() int {
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

func (a *Manager) IsBootstrapped() bool {
	return a.IsUsersBootstrapped() && a.IsSettingsBootstrapped() && a.IsStateBootstrapped()
}

func (a *Manager) IsUsersBootstrapped() bool {
	return a.users != nil && a.sessions != nil && a.tokens != nil
}

func (a *Manager) IsStateBootstrapped() bool {
	return a.state != nil
}

func (a *Manager) IsSettingsBootstrapped() bool {
	return a.cm_settings != nil
}

func (a *Manager) ResetBootstrapState() error {
	a.logger = nil

	a.sessions = nil
	a.users = nil
	a.state = nil
	a.tokens = nil

	// We do this last since it can err
	// TODO: close all dbs
	if a.cm_db != nil {
		if err := a.cm_db.Close(); err != nil {
			return err
		}
	}

	return nil
}

// Logger returns the default app logger.
//
// If the application is not bootstrapped yet, fallbacks to slog.Default().
func (app *Manager) Logger() *slog.Logger {
	if app.logger == nil {
		return slog.Default()
	}
	return app.logger
}

// IsDev returns true if the application is running in development mode.
func (app *Manager) IsDev() bool {
	return app.isDev
}

func (app *Manager) Sessions() *sessions.SessionManager {
	return app.sessions
}

func (app *Manager) Users() *users.UserManager {
	return app.users
}

func (app *Manager) DataDir() string {
	return app.dataDir
}

func (a *Manager) CMSettings() *models.Settings {
	return a.cm_settings
}

// INFO: every init function must make sure of it's own dependencies
func (app *Manager) InitLogger() error {
	// TODO: create a handler, to write logs to the db
	app.logger = slog.Default()
	return nil
}

func (app *Manager) CreateDB(dir string, file string, maxopenconns int, maxidleconns int) (*db.DB, error) {
	p := filepath.Join(dir, file)
	db, err := db.New(p, maxopenconns, maxidleconns)
	if err != nil {
		return nil, err
	}

	if app.isDev {
		db.ConnectLogger()
	}

	return db, nil
}

func (a *Manager) InitDataStore(db *db.DB, tablename string, idfield string) error {
	ds, err := datastore.New(db, tablename, idfield)
	if err != nil {
		return err
	}
	a.state = ds
	return nil
}

func (a *Manager) InitSettings(dsm *datastore.DataStoreManager, key string) error {
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

	a.cm_settings = sets
	return nil
}

func (a *Manager) InitSessions(db *db.DB, stn, utn, idfield string, lsessexp, ssessexp int, sets *models.Settings) error {
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

func (a *Manager) InitUsers(db *db.DB, utn, idfield string, uexp int, sets *models.Settings) error {
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

func (a *Manager) InitTokens(db *db.DB, atn, utn, idfield string, lressexp, sressexp int) error {
	tm, err := accesstokens.New(db, atn, utn, idfield, lressexp, sressexp)
	if err != nil {
		return err
	}
	a.tokens = tm
	return nil
}
