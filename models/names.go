package models

const (
	// ROUTES
	ASSET_ROUTE    = "assets"
	CDN_BASE_ROUTE = "cdn"
	FILE_ROUTE     = "files"

	VERSION               = "0.1.0"
	STORE_KEY_SETUP_STATE = "setup"
	STATIC_FILEPATH       = "./frontend/assets"
	ROUTES_FILEPATH       = "./frontend/routes"

	DEFAULT_DATA_MAX_OPEN_CONNS int = 120
	DEFAULT_DATA_MAX_IDLE_CONNS int = 20
	DEFAULT_LOGS_MAX_OPEN_CONNS int = 10
	DEFAULT_LOGS_MAX_IDLE_CONNS int = 2

	DEFAULT_LOCAL_STORAGE_DIR string = "storage"
	DEFAULT_BACKUPS_DIR       string = "backups"

	// TODO: true for dev purposes
	DEFAULT_DEV_MODE      bool   = true
	DEFAULT_TEST_DATA_DIR string = "test_data"
	DEFAULT_DATA_DIR      string = "cm_data"
	DEFAULT_DATA_FILE     string = "data.db"

	DEFAULT_SESSIONS_TABLE      string = "__sessions"
	DEFAULT_ACCESS_TOKENS_TABLE string = "__access_tokens"
	DEFAULT_USERS_TABLE         string = "__users"
	DEFAULT_MIGRATIONS_TABLE    string = "__migrations"
	DEFAULT_DATASTORE_TABLE     string = "__datastore"
	DEFAULT_ID_FIELD            string = "id"

	DATASTORE_SETTINGS_KEY string = "sets"

	DEFAULT_USER_EXPIRATION          int = 60 * 60 * 24 * (365 * 10) // ~10 years
	DEFAULT_SHORT_SESSION_EXPIRATION int = 60 * 60 * 2               // 2 hours
	DEFAULT_CSRF_EXPIRATION          int = 60 * 60 * 24              // 1 day
	DEFAULT_LONG_SESSION_EXPIRATION  int = 60 * 60 * 24 * 30         // 30 days

	DEFAULT_LONG_RESOURCE_SESSION_EXPIRATION  int = 60 * 60 * 24 * 7 // 7 days
	DEFAULT_SHORT_RESOURCE_SESSION_EXPIRATION int = 60 * 60 * 6      // 6 hours
)
