package users

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"git.terah.dev/imterah/hermes/apiclient"
	"git.terah.dev/imterah/hermes/frontend/config"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

func CreateUserCommand(cCtx *cli.Context) error {
	configPath := cCtx.String("config-path")

	var configContents *config.Config

	_, err := os.Stat(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configContents = &config.Config{}
		} else {
			return fmt.Errorf("failed to get configuration file information: %s", err.Error())
		}
	} else {
		configContents, err = config.ReadAndParseConfig(configPath)

		if err != nil {
			return fmt.Errorf("failed to read and parse configuration file: %s", err.Error())
		}
	}

	username := cCtx.String("username")

	if username == "" {
		if configContents.Username == "" {
			return fmt.Errorf("username not specified and username is not in the configuration file")
		}

		username = configContents.Username
	}

	var password string

	if cCtx.Bool("ask-password") {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Print("\n")

		if err != nil {
			return fmt.Errorf("failed to read password from console: %s", err.Error())
		}

		password = string(passwordBytes)
	} else {
		password = cCtx.String("password")

		if password == "" {
			return fmt.Errorf("password is not specified and password asking is not enabled")
		}
	}

	var serverURL string

	if cCtx.String("server-url") == "" {
		if configContents.APIPath == "" {
			return fmt.Errorf("server URL not specified and server URL is not in the configuration file")
		}

		serverURL = configContents.APIPath
	} else {
		serverURL = cCtx.String("server-url")
	}

	fullName := cCtx.String("full-name")
	email := cCtx.String("email")
	isBot := cCtx.Bool("user-is-bot")

	log.Info("Creating user...")

	api := &apiclient.HermesAPIClient{
		URL: serverURL,
	}

	refreshToken, err := api.UserCreate(fullName, username, email, password, isBot)

	if err != nil {
		return fmt.Errorf("failed to create user: %s", err.Error())
	}

	log.Info("Successfully created user.")

	if cCtx.Bool("do-not-save-configuration") {
		return nil
	}

	configContents.Username = username
	configContents.RefreshToken = refreshToken
	configContents.APIPath = serverURL

	data, err := yaml.Marshal(configContents)

	if err != nil {
		return fmt.Errorf("failed to marshal configuration data: %s", err.Error())
	}

	os.WriteFile(configPath, data, 0644)

	return nil
}
