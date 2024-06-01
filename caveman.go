package caveman

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Simon-Martens/caveman/app"
	"github.com/Simon-Martens/caveman/cmd"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/list"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Caveman defines a Caveman apppplication
// It embeds the App methods (providing all kinds of low-level access)
// But here are some higher-level methods given for convenience. Also,
// the user can give a custom settings struct to run the app with, that
// optionally can be saved as the most recent settings into the DB.
type Caveman struct {
	*app.App

	RootCmd         *cobra.Command
	StartupSettings models.Config
}

// Creates a new Caveman instance with either the
// - latest settings from the DB
// - or the default settings if no settings are found.
func New() *Caveman {
	return NewWithSettings(models.Config{})
}

// NewWithSettings creates a new Caveman instance with the provided config.
//
// Note that the application will not be initialized/bootstrapped yet,
// aka. DB connections, migrations, app settings, etc. will not be accessible.
// Everything will be initialized when [Start()] is executed.
// If you want to initialize the application before calling [Start()],
// then you'll have to manually call [Bootstrap()].
func NewWithSettings(settings models.Config) *Caveman {
	baseDir, dev := inspectRuntime()
	if settings.DataDir == "" {
		settings.DataDir = filepath.Join(baseDir, "cm_data")
	}

	cm := &Caveman{
		RootCmd: &cobra.Command{
			Use:     filepath.Base(os.Args[0]),
			Short:   "Caveman CLI",
			Version: models.VERSION,
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		},
		StartupSettings: settings,
	}

	cm.RootCmd.SetErr(newErrWriter())

	// if dev is false up until this point, then the default dev mode can overwrite
	cm.eagerParseFlags(dev || models.DEFAULT_DEV_MODE)

	cm.App = app.New(cm.StartupSettings)

	cm.RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return cm
}

// Start starts the application, aka. registers the default system
// commands (serve, migrate, version) and executes cm.RootCmd.
func (cm *Caveman) Start() error {
	cm.RootCmd.AddCommand(cmd.CmdServe(cm.App))
	return cm.Execute()
}

func (cm *Caveman) Execute() error {
	if !cm.skipBootstrap() {
		if err := cm.Bootstrap(); err != nil {
			return err
		}
	}

	done := make(chan bool, 1)

	// gracefully shutdown the application on interrupt signals
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	go func() {
		cm.Logger().Info("Caveman CLI v " + models.VERSION)
		cm.RootCmd.Execute()

		done <- true
	}()

	<-done

	// trigger cleanups
	return cm.Terminate()
}

func (cm *Caveman) eagerParseFlags(dev bool) error {
	cm.RootCmd.PersistentFlags().StringVar(
		&cm.StartupSettings.DataDir,
		"dir",
		models.DEFAULT_DATA_DIR_NAME,
		"the Caveman data directory",
	)

	cm.RootCmd.PersistentFlags().BoolVar(
		&cm.StartupSettings.Dev,
		"dev",
		dev,
		"print logs and sql statements to the console",
	)

	return cm.RootCmd.ParseFlags(os.Args[1:])
}

// skicmootstrap eagerly checks if the app should skip the bootstrap process:
// - already bootstrapped
// - is unknown command
// - is the default help command
// - is the default version command
func (cm *Caveman) skipBootstrap() bool {
	flags := []string{
		"-h",
		"--help",
		"-v",
		"--version",
	}

	if cm.IsBootstrapped() {
		return true // already bootstrapped
	}

	cmd, _, err := cm.RootCmd.Find(os.Args[1:])
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
