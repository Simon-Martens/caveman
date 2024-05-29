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
	"github.com/Simon-Martens/caveman/tools/list"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Version of Caveman
var Version = "(untracked)"

// Caveman defines a Caveman app launcher.
//
// It ebends the App + startup config and all of the app methods
// can be accessed directly through the instance (eg. Caveman.DataDir()).
type Caveman struct {
	*app.App

	devFlag     bool
	dataDirFlag string

	// RootCmd is the main console command
	RootCmd *cobra.Command
}

// Config is the Caveman initialization config struct.
type Config struct {
	DefaultDev     bool
	DefaultDataDir string // if not set, it will fallback to "./cm_data"
}

// New creates a new Caveman instance with the default configuration.
// Use [NewWithConfig()] if you want to provide a custom configuration.
//
// Note that the application will not be initialized/bootstrapped yet,
// aka. DB connections, migrations, app settings, etc. will not be accessible.
// Everything will be initialized when [Start()] is executed.
// If you want to initialize the application before calling [Start()],
// then you'll have to manually call [Bootstrap()].
func New() *Caveman {
	_, isUsingGoRun := inspectRuntime()

	return NewWithConfig(Config{
		DefaultDev: isUsingGoRun,
	})
}

// NewWithConfig creates a new Caveman instance with the provided config.
//
// Note that the application will not be initialized/bootstrapped yet,
// aka. DB connections, migrations, app settings, etc. will not be accessible.
// Everything will be initialized when [Start()] is executed.
// If you want to initialize the application before calling [Start()],
// then you'll have to manually call [Bootstrap()].
func NewWithConfig(config Config) *Caveman {
	// initialize a default data directory based on the executable baseDir
	if config.DefaultDataDir == "" {
		baseDir, _ := inspectRuntime()
		config.DefaultDataDir = filepath.Join(baseDir, "cm_data")
	}

	cm := &Caveman{
		RootCmd: &cobra.Command{
			Use:     filepath.Base(os.Args[0]),
			Short:   "Caveman CLI",
			Version: Version,
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
			// no need to provide the default cobra completion command
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		},
		devFlag:     config.DefaultDev,
		dataDirFlag: config.DefaultDataDir,
	}

	// replace with a colored stderr writer
	cm.RootCmd.SetErr(newErrWriter())

	// parse base flags
	// (errors are ignored, since the full flags parsing happens on Execute())
	cm.eagerParseFlags(&config)

	// initialize the app instance
	app, err := app.New(cm.devFlag, cm.dataDirFlag)

	if err != nil {
		cm.RootCmd.Println("Error initializing the application:", err)
		os.Exit(1)
	}

	cm.App = app

	// hide the default help command (allow only `--help` flag)
	cm.RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return cm
}

// Start starts the application, aka. registers the default system
// commands (serve, migrate, version) and executes cm.RootCmd.
func (cm *Caveman) Start() error {
	// register system commands
	cm.RootCmd.AddCommand(cmd.CmdServe(cm.App))

	return cm.Execute()
}

// Execute initializes the application (if not already) and executes
// the cm.RootCmd with graceful shutdown support.
//
// This method differs from cm.Start() by not registering the default
// system commands!
func (cm *Caveman) Execute() error {
	if !cm.skicmootstrap() {
		if err := cm.Bootstrap(); err != nil {
			return err
		}
	}

	done := make(chan bool, 1)

	// listen for interrupt signal to gracefully shutdown the application
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		done <- true
	}()

	// execute the root command
	go func() {
		// note: leave to the commands to decide whether to print their error
		cm.RootCmd.Execute()

		done <- true
	}()

	<-done

	// trigger cleanups
	return cm.Terminate()
}

// eagerParseFlags parses the global app flags before calling cm.RootCmd.Execute().
// so we can have all Caveman flags ready for use on initialization.
func (cm *Caveman) eagerParseFlags(config *Config) error {
	cm.RootCmd.PersistentFlags().StringVar(
		&cm.dataDirFlag,
		"dir",
		config.DefaultDataDir,
		"the Caveman data directory",
	)

	cm.RootCmd.PersistentFlags().BoolVar(
		&cm.devFlag,
		"dev",
		config.DefaultDev,
		"enable dev mode, aka. printing logs and sql statements to the console",
	)

	return cm.RootCmd.ParseFlags(os.Args[1:])
}

// skicmootstrap eagerly checks if the app should skip the bootstrap process:
// - already bootstrapped
// - is unknown command
// - is the default help command
// - is the default version command
func (cm *Caveman) skicmootstrap() bool {
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
