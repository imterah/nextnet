package main

import (
	"os"
	"path"

	"git.terah.dev/imterah/hermes/frontend/commands/users"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

func main() {
	logLevel := os.Getenv("HERMES_LOG_LEVEL")

	if logLevel != "" {
		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)

		case "info":
			log.SetLevel(log.InfoLevel)

		case "warn":
			log.SetLevel(log.WarnLevel)

		case "error":
			log.SetLevel(log.ErrorLevel)

		case "fatal":
			log.SetLevel(log.FatalLevel)
		}
	}

	configDir, err := os.UserConfigDir()

	if err != nil {
		log.Fatalf("Failed to get configuration directory: %s", err.Error())
	}

	app := &cli.App{
		Name:  "hermcli",
		Usage: "client for Hermes -- port forwarding across boundaries",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-path",
				Aliases: []string{"config", "cp", "c"},
				Value:   path.Join(configDir, "hermcli.yml"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "login",
				Usage:   "log in to the API",
				Action:  users.GetRefreshTokenCommand,
				Aliases: []string{"l"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "username",
						Aliases: []string{"user", "u"},
						Usage:   "username to authenticate as",
					},
					&cli.StringFlag{
						Name:    "password",
						Aliases: []string{"pass", "p"},
						Usage:   "password to authenticate with",
					},
					&cli.StringFlag{
						Name:    "server-url",
						Aliases: []string{"server", "s"},
						Usage:   "URL of the server to authenticate with",
					},
					&cli.BoolFlag{
						Name:    "ask-password",
						Aliases: []string{"ask-pass", "ap"},
						Usage:   "asks you the password to authenticate with",
					},
				},
			},
			{
				Name:    "users",
				Usage:   "user management commands",
				Aliases: []string{"u"},
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "create a user",
						Action:  users.CreateUserCommand,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "full-name",
								Aliases:  []string{"name", "n"},
								Usage:    "full name for the user",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "username",
								Aliases:  []string{"user", "us"},
								Usage:    "username to give the user",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "email",
								Aliases:  []string{"e"},
								Usage:    "email to give the user",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "password",
								Aliases: []string{"pass", "p"},
								Usage:   "password to give the user",
							},
							&cli.StringFlag{
								Name:    "server-url",
								Aliases: []string{"server", "s"},
								Usage:   "URL of the server to connect with",
							},
							&cli.BoolFlag{
								Name:    "ask-password",
								Aliases: []string{"ask-pass", "ap"},
								Usage:   "asks you the password to give the user",
							},
							&cli.BoolFlag{
								Name:    "user-is-bot",
								Aliases: []string{"user-bot", "ub", "u"},
								Usage:   "if set, makes the user flagged as a bot",
							},
							&cli.BoolFlag{
								Name:    "do-not-save-configuration",
								Aliases: []string{"no-save", "ns"},
								Usage:   "doesn't save the authenticated user credentials",
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
