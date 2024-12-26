package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"git.terah.dev/imterah/hermes/api/dbcore"
	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

// Data structures
type BackupBackend struct {
	ID uint `json:"id" validate:"required"`

	Name              string  `json:"name" validate:"required"`
	Description       *string `json:"description"`
	Backend           string  `json:"backend" validate:"required"`
	BackendParameters string  `json:"connectionDetails" validate:"required"`
}

type BackupProxy struct {
	ID        uint `json:"id" validate:"required"`
	BackendID uint `json:"destProviderID" validate:"required"`

	Name            string  `json:"name" validate:"required"`
	Description     *string `json:"description"`
	Protocol        string  `json:"protocol" validate:"required"`
	SourceIP        string  `json:"sourceIP" validate:"required"`
	SourcePort      uint16  `json:"sourcePort" validate:"required"`
	DestinationPort uint16  `json:"destPort" validate:"required"`
	AutoStart       bool    `json:"enabled" validate:"required"`
}

type BackupPermission struct {
	ID uint `json:"id" validate:"required"`

	PermissionNode string `json:"permission" validate:"required"`
	HasPermission  bool   `json:"has" validate:"required"`
	UserID         uint   `json:"userID" validate:"required"`
}

type BackupUser struct {
	ID uint `json:"id" validate:"required"`

	Email    string  `json:"email" validate:"required"`
	Username *string `json:"username"`
	Name     string  `json:"name" validate:"required"`
	Password string  `json:"password" validate:"required"`
	IsBot    *bool   `json:"isRootServiceAccount"`

	Token *string `json:"rootToken" validate:"required"`
}

type BackupData struct {
	Backends    []*BackupBackend    `json:"destinationProviders" validate:"required"`
	Proxies     []*BackupProxy      `json:"forwardRules" validate:"required"`
	Permissions []*BackupPermission `json:"allPermissions" validate:"required"`
	Users       []*BackupUser       `json:"users" validate:"required"`
}

// From https://stackoverflow.com/questions/54461423/efficient-way-to-remove-all-non-alphanumeric-characters-from-large-text
func stripAllAlphanumeric(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func backupRestoreEntrypoint(cCtx *cli.Context) error {
	log.Info("Decompressing backup...")

	backupFile, err := os.Open(cCtx.String("backup-path"))

	if err != nil {
		return fmt.Errorf("failed to open backup: %s", err.Error())
	}

	reader, err := gzip.NewReader(backupFile)
	backupDataBytes, err := io.ReadAll(reader)

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Decompressed backup. Cleaning up...")

	err = reader.Close()

	if err != nil {
		return fmt.Errorf("failed to close Gzip reader: %s", err.Error())
	}

	err = backupFile.Close()

	if err != nil {
		return fmt.Errorf("failed to close backup: %s", err.Error())
	}

	log.Info("Parsing backup into internal structures...")

	backupData := &BackupData{}

	err = json.Unmarshal(backupDataBytes, backupData)

	if err != nil {
		return fmt.Errorf("failed to parse backup: %s", err.Error())
	}

	if err := validator.New().Struct(backupData); err != nil {
		return fmt.Errorf("failed to validate backup: %s", err.Error())
	}

	log.Warn("!! WARNING !!")
	log.Warn("This will attempt to permanently wipe the old database. The backup will not be deleted, however, caution is still advised.")
	log.Warn("Continuing in 5 seconds...")

	time.Sleep(5 * time.Second)

	log.Info("Wiping database...")

	databaseBackend := os.Getenv("HERMES_DATABASE_BACKEND")

	switch databaseBackend {
	case "sqlite":
		filePath := os.Getenv("HERMES_SQLITE_FILEPATH")

		if filePath == "" {
			return fmt.Errorf("sqlite database file not specified (missing HERMES_SQLITE_FILEPATH)")
		}

		err := os.Remove(filePath)

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to delete sqlite database: %s", err.Error())
		}
	case "postgresql":
		// FIXME(imterah): Maybe make this not required?
		postgresDB := os.Getenv("HERMES_MIGRATE_POSTGRES_DATABASE")

		if postgresDB == "" {
			return fmt.Errorf("postgres migration DB is not specified (we don't parse the DSN to save space) (missing HERMES_MIGRATE_POSTGRES_DATABASE)")
		}

		postgresDSN := os.Getenv("HERMES_POSTGRES_DSN")

		if postgresDSN == "" {
			return fmt.Errorf("postgres DSN not specified (missing HERMES_POSTGRES_DSN)")
		}

		log.Info("Connecting to database...")

		db, err := sql.Open("postgres", postgresDSN)

		if err != nil {
			return fmt.Errorf("failed to connect to database: %s", err.Error())
		}

		log.Info("Dropping database...")

		_, err = db.Query("DROP DATABASE ?", postgresDB)

		if err != nil {
			return fmt.Errorf("failed to drop database: %s", err.Error())
		}

		log.Info("Closing database connection...")

		err = db.Close()

		if err != nil {
			return fmt.Errorf("failed to close database connection: %s", err.Error())
		}
	case "":
		return fmt.Errorf("no database backend specified in environment variables (missing HERMES_DATABASE_BACKEND)")
	default:
		return fmt.Errorf("unknown database backend specified: %s", os.Getenv(databaseBackend))
	}

	log.Info("Reinitializing database and opening it...")

	err = dbcore.InitializeDatabase(&gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}

	log.Info("Running database migrations...")

	if err := dbcore.DoDatabaseMigrations(dbcore.DB); err != nil {
		return fmt.Errorf("Failed to run database migrations: %s", err)
	}

	log.Info("Restoring database...")
	bestEffortOwnerUIDFromBackup := -1

	log.Info("Attempting to find user to use as owner of resources...")

	for _, user := range backupData.Users {
		foundUser := false
		failedAdministrationCheck := false

		for _, permission := range backupData.Permissions {
			if permission.UserID != user.ID {
				continue
			}

			foundUser = true

			if !strings.HasPrefix(permission.PermissionNode, "routes.") && permission.PermissionNode != "permissions.see" && !permission.HasPermission {
				log.Infof("User with email '%s' and ID of '%d' failed administration check (lacks all permissions required). Attempting to find better user", user.Email, user.ID)
				failedAdministrationCheck = true

				break
			}
		}

		if !foundUser {
			log.Warnf("User with email '%s' and ID of '%d' lacks any permissions!", user.Email, user.ID)
			continue
		}

		if failedAdministrationCheck {
			continue
		}

		log.Infof("Using user with email '%s', and ID of '%d'", user.Email, user.ID)
		bestEffortOwnerUIDFromBackup = int(user.ID)

		break
	}

	if bestEffortOwnerUIDFromBackup == -1 {
		log.Warnf("Could not find Administrative level user to use as the owner of resources. Using user with email '%s', and ID of '%d'", backupData.Users[0].Email, backupData.Users[0].ID)
		bestEffortOwnerUIDFromBackup = int(backupData.Users[0].ID)
	}

	var bestEffortOwnerUID uint

	for _, user := range backupData.Users {
		log.Debugf("Migrating user with email '%s' and ID of '%d'", user.Email, user.ID)
		tokens := make([]dbcore.Token, 0)
		permissions := make([]dbcore.Permission, 0)

		if user.Token != nil {
			tokens = append(tokens, dbcore.Token{
				Token:          *user.Token,
				DisableExpiry:  true,
				CreationIPAddr: "127.0.0.1", // We don't know the creation IP address...
			})
		}

		for _, permission := range backupData.Permissions {
			if permission.UserID != user.ID {
				continue
			}

			permissions = append(permissions, dbcore.Permission{
				PermissionNode: permission.PermissionNode,
				HasPermission:  permission.HasPermission,
			})
		}

		username := ""

		if user.Username == nil {
			username = strings.ToLower(stripAllAlphanumeric(user.Email))
			log.Warnf("User with ID of '%d' doesn't have a username. Derived username from email is '%s' (email is '%s')", user.ID, username, user.Email)
		} else {
			username = *user.Username
		}

		userDatabase := &dbcore.User{
			Email:    user.Email,
			Username: username,
			Name:     user.Name,
			Password: user.Password,
			IsBot:    user.IsBot,

			Tokens:      tokens,
			Permissions: permissions,
		}

		if err := dbcore.DB.Create(userDatabase).Error; err != nil {
			log.Errorf("Failed to create user: %s", err.Error())
		}

		if uint(bestEffortOwnerUIDFromBackup) == user.ID {
			bestEffortOwnerUID = userDatabase.ID
		}
	}

	for _, backend := range backupData.Backends {
		log.Debugf("Migrating backend ID '%d' with name '%s'", backend.ID, backend.Name)

		backendDatabase := &dbcore.Backend{
			UserID:            bestEffortOwnerUID,
			Name:              backend.Name,
			Description:       backend.Description,
			Backend:           backend.Backend,
			BackendParameters: backend.BackendParameters,
		}

		if err := dbcore.DB.Create(backendDatabase).Error; err != nil {
			log.Errorf("Failed to create backend: %s", err.Error())
		}

		log.Debugf("Migrating proxies for backend ID '%d'", backend.ID)

		for _, proxy := range backupData.Proxies {
			if proxy.BackendID != backend.ID {
				continue
			}

			log.Debugf("Migrating proxy ID '%d' with name '%s'", proxy.ID, proxy.Name)

			proxyDatabase := &dbcore.Proxy{
				BackendID: backendDatabase.ID,
				UserID:    bestEffortOwnerUID,

				Name:            proxy.Name,
				Description:     proxy.Description,
				Protocol:        proxy.Protocol,
				SourceIP:        proxy.SourceIP,
				SourcePort:      proxy.SourcePort,
				DestinationPort: proxy.DestinationPort,
				AutoStart:       proxy.AutoStart,
			}

			if err := dbcore.DB.Create(proxyDatabase).Error; err != nil {
				log.Errorf("Failed to create proxy: %s", err.Error())
			}
		}
	}

	log.Info("Successfully upgraded to Hermes from NextNet.")

	return nil
}
