// Package migratecmd adds a new "migrate" command support to a PocketBase instance.
//
// It also comes with automigrations support and templates generation
// (both for JS and GO migration files).
//
// Example usage:
//
//	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
//		TemplateLang: migratecmd.TemplateLangJS, // default to migratecmd.TemplateLangGo
//		Automigrate:  true,
//		Dir:          "/custom/migrations/dir", // optional template migrations path; default to "pb_migrations" (for JS) and "migrations" (for Go)
//	})
//
//	Note: To allow running JS migrations you'll need to enable first
//	[jsvm.MustRegister()].
package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/Simon-Martens/caveman/app"
	"github.com/Simon-Martens/caveman/migrations"
	"github.com/Simon-Martens/caveman/tools/migration"
	"github.com/spf13/cobra"
)

// MustRegister registers the migratecmd plugin to the provided app instance
// and panic if it fails.
//
// Example usage:
//
//	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{})
func MustRegister(app app.App, rootCmd *cobra.Command, dir string) {
	if err := Register(app, rootCmd, dir); err != nil {
		panic(err)
	}
}

// Register registers the migratecmd plugin to the provided app instance.
func Register(app app.App, rootCmd *cobra.Command, dir string) error {
	p := &plugin{app: app, dir: dir}

	if dir == "" {
		dir = filepath.Join(p.app.DataDir(), "../migrations")
	}

	// attach the migrate command
	if rootCmd != nil {
		rootCmd.AddCommand(p.createCommand())
	}

	return nil
}

type plugin struct {
	app app.App
	dir string
}

func (p *plugin) createCommand() *cobra.Command {
	const cmdDesc = `Supported arguments are:
- up            - runs all available migrations
- down [number] - reverts the last [number] applied migrations
- create name   - creates new blank migration template file
- history-sync  - ensures that the _migrations history table doesn't have references to deleted migration files
`

	command := &cobra.Command{
		Use:          "migrate",
		Short:        "Executes app DB migration scripts",
		Long:         cmdDesc,
		ValidArgs:    []string{"up", "down", "create"},
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			cmd := ""
			if len(args) > 0 {
				cmd = args[0]
			}

			switch cmd {
			case "create":
				if _, err := p.migrateCreateHandler("", args[1:], true); err != nil {
					return err
				}
			default:
				runner, err := migration.NewRunner(p.app.DB().NonConcurrentDB(), migrations.AppMigrations)
				if err != nil {
					return err
				}

				if err := runner.Run(args...); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return command
}

func (p *plugin) migrateCreateHandler(template string, args []string, interactive bool) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("Missing migration file name")
	}

	name := args[0]
	dir := p.dir

	filename := fmt.Sprintf("%d_%s.%s", time.Now().Unix(), name, "go")

	resultFilePath := path.Join(dir, filename)

	if interactive {
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Do you really want to create migration %q?", resultFilePath),
		}
		survey.AskOne(prompt, &confirm)
		if !confirm {
			fmt.Println("The command has been cancelled")
			return "", nil
		}
	}

	// get default create template
	if template == "" {
		t, templateErr := p.goBlankTemplate()
		if templateErr != nil {
			return "", fmt.Errorf("Failed to resolve create template: %v\n", templateErr)
		}
		template = t
	}

	// ensure that the migrations dir exist
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// save the migration file
	if err := os.WriteFile(resultFilePath, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("Failed to save migration file %q: %v\n", resultFilePath, err)
	}

	if interactive {
		fmt.Printf("Successfully created file %q\n", resultFilePath)
	}

	return filename, nil
}

func (p *plugin) goBlankTemplate() (string, error) {
	const template = `package %s

import (
	"github.com/pocketbase/dbx"
	m "github.com/Simon-Martens/caveman/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		// add up queries...

		return nil
	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
`

	return fmt.Sprintf(template, filepath.Base(p.config.Dir)), nil
}
