package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Simon-Martens/caveman/app"
	"github.com/spf13/cobra"
)

const (
	VERSION = "0.1.0"
)

func RootCmd(app *app.App) *cobra.Command {
	command := &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Short:   "caveman CLI",
		Version: VERSION,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	command.AddCommand(CmdServe(app))
	return command
}

func Execute(app *app.App) {
	if err := RootCmd(app).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
