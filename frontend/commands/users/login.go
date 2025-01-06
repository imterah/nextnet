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

func GetRefreshTokenCommand(cCtx *cli.Context) error {
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

	var username string
	var password string

	if cCtx.String("username") == "" {
		if configContents.Username == "" {
			return fmt.Errorf("username not specified and username is not in the configuration file")
		}

		username = configContents.Username
	} else {
		username = cCtx.String("username")
	}

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

	serverURL := cCtx.String("server-url")
	log.Info("Authenticating with API...")

	api := &apiclient.HermesAPIClient{
		URL: serverURL,
	}

	refreshToken, err := api.UserGetRefreshToken(&username, nil, password)

	if err != nil {
		return fmt.Errorf("failed to authenticate with the API: %s", err.Error())
	}

	configContents.Username = username
	configContents.RefreshToken = refreshToken
	configContents.APIPath = serverURL

	log.Info("Writing configuration file...")
	data, err := yaml.Marshal(configContents)

	if err != nil {
		return fmt.Errorf("failed to marshal configuration data: %s", err.Error())
	}

	log.Infof("config path: %s", configPath)

	os.WriteFile(configPath, data, 0644)

	return nil
}
