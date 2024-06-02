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

	DEFAULT_LOCAL_STORAGE_DIR_NAME string = "storage"
	DEFAULT_BACKUPS_DIR_NAME       string = "backups"

	// TODO: true for dev purposes
	DEFAULT_DEV_MODE           bool   = true
	DEFAULT_TEST_DATA_DIR_NAME string = "test_data"
	DEFAULT_DATA_DIR_NAME      string = "cm_data"
	DEFAULT_DATA_FILE_NAME     string = "data.db"

	DEFAULT_SESSIONS_TABLE_NAME   string = "__sessions"
	DEFAULT_USERS_TABLE_NAME      string = "__users"
	DEFAULT_MIGRATIONS_TABLE_NAME string = "__migrations"
	DEFAULT_ID_FIELD              string = "id"

	DEFAULT_USER_EXPIRATION          int = 60 * 60 * 24 * (365 * 10) // ~10 years
	DEFAULT_SHORT_SESSION_EXPIRATION int = 60 * 60 * 2               // 2 hours
	DEFAULT_LONG_SESSION_EXPIRATION  int = 60 * 60 * 24 * 30         // 30 days
)
