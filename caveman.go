package caveman

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Simon-Martens/caveman/manager"
	"github.com/Simon-Martens/caveman/migrations"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/list"
	"github.com/Simon-Martens/caveman/tools/migration"
	"github.com/fatih/color"
	"github.com/pocketbase/dbx"
	"github.com/spf13/cobra"
)

// TODO: we don't need to parse command line flags since we are not a command line
// application. Maybe just keep some cobra commands in the cmd/ directory to be
// available to applications.

// Caveman defines a Caveman apppplication
// It embeds the App methods (providing all kinds of low-level access)
// But here are some higher-level methods given for convenience. Also,
// the user can give a custom settings struct to run the app with, that
// optionally can be saved as the most recent settings into the DB.
type Caveman struct {
	*manager.Manager
	StartupSettings models.Config
}

// Creates a new Caveman instance with either the
// - latest settings from the DB
// - or the default settings if no settings are found.
func New() *Caveman {
	return NewWithSettings(models.Config{})
}

// NewWithSettings creates a new Caveman with the provided config.
// Note that the database manager will not be initialized/bootstrapped yet,
// aka. DB connections, migrations, app settings, etc. will not be accessible.
// Everything will be initialized when [Bootstrap()] is called.
func NewWithSettings(settings models.Config) *Caveman {
	baseDir, dev := inspectRuntime()
	if settings.DataDir == "" {
		settings.DataDir = filepath.Join(baseDir, "cm_data")
	}

	// We force dev mode if run with go run
	if dev {
		settings.Dev = true
	}

	cm := &Caveman{StartupSettings: settings}
	cm.Manager = manager.New(cm.StartupSettings)

	return cm
}

// Given a cobra command, this parses the caveman relevant flags.
func (cm *Caveman) ParseFlags(dev bool, rootCmd *cobra.Command) error {
	rootCmd.PersistentFlags().StringVar(
		&cm.StartupSettings.DataDir,
		"dir",
		models.DEFAULT_DATA_DIR,
		"the Caveman data directory",
	)

	rootCmd.PersistentFlags().BoolVar(
		&cm.StartupSettings.Dev,
		"dev",
		dev,
		"print logs and sql statements to the console",
	)

	return rootCmd.ParseFlags(os.Args[1:])
}

// skicmootstrap eagerly checks if the app should skip the bootstrap process:
// - already bootstrapped
// plus, if given a cobra command:
// - is unknown command
// - is the default help command
// - is the default version command
func (cm *Caveman) skipBootstrap(cmd *cobra.Command) bool {
	flags := []string{
		"-h",
		"--help",
		"-v",
		"--version",
	}

	if cm.IsBootstrapped() {
		return true // already bootstrapped
	}

	if cmd != nil {
		cmd, _, err := cmd.Find(os.Args[1:])
		if err != nil {
			return true // unknown command
		}

		for _, arg := range os.Args {
			if !list.ExistInSlice(arg, flags) {
				continue
			}

			// ensure that there is no user defined flag with the same name/shorthand
			trimmed := strings.TrimLeft(arg, "-")
			if len(trimmed) > 1 && cmd.Flags().Lookup(trimmed) == nil {
				return true
			}
			if len(trimmed) == 1 && cmd.Flags().ShorthandLookup(trimmed) == nil {
				return true
			}
		}
	}

	return false
}

// inspectRuntime tries to find the base executable directory and how it was run.
func inspectRuntime() (baseDir string, withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// probably ran with go run
		withGoRun = true
		baseDir, _ = os.Getwd()
	} else {
		// probably ran with go build
		withGoRun = false
		baseDir = filepath.Dir(os.Args[0])
	}
	return
}

// newErrWriter returns a red colored stderr writter.
func newErrWriter() *coloredWriter {
	return &coloredWriter{
		w: os.Stderr,
		c: color.New(color.FgRed),
	}
}

// coloredWriter is a small wrapper struct to construct a [color.Color] writter.
type coloredWriter struct {
	w io.Writer
	c *color.Color
}

// Write writes the p bytes using the colored writer.
func (colored *coloredWriter) Write(p []byte) (n int, err error) {
	colored.c.SetWriter(colored.w)
	defer colored.c.UnsetWriter(colored.w)

	return colored.c.Print(string(p))
}

type migrationsConnection struct {
	DB             *dbx.DB
	MigrationsList migration.MigrationsList
}

func RunMigrations(app manager.Manager) error {
	connections := []migrationsConnection{
		{
			DB:             app.DB().NonConcurrentDB(),
			MigrationsList: migrations.AppMigrations,
		},
	}

	for _, c := range connections {
		runner, err := migration.NewRunner(c.DB, c.MigrationsList)
		if err != nil {
			return err
		}

		if _, err := runner.Up(); err != nil {
			return err
		}
	}

	return nil
}
