package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Simon-Martens/caveman/app"
	"github.com/Simon-Martens/caveman/server"
	"github.com/spf13/cobra"
)

func CmdServe(app *app.App) *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			if err := server.Serve(ctx, app); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		},
	}
	return serveCmd
}
