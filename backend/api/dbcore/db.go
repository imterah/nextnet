package dbcore

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Backend struct {
	gorm.Model

	UserID uint

	Name              string
	Description       *string
	Backend           string
	BackendParameters string

	Proxies []Proxy
}

type Proxy struct {
	gorm.Model

	BackendID uint
	UserID    uint

	Name            string
	Description     *string
	Protocol        string
	SourceIP        string
	SourcePort      uint16
	DestinationPort uint16
	AutoStart       bool
}

type Permission struct {
	gorm.Model

	PermissionNode string
	HasPermission  bool
	UserID         uint
}

type Token struct {
	gorm.Model

	UserID uint

	Token          string
	DisableExpiry  bool
	CreationIPAddr string
}

type User struct {
	gorm.Model

	Email    string `gorm:"unique"`
	Username string `gorm:"unique"`
	Name     string
	Password string
	IsBot    *bool

	Permissions   []Permission
	OwnedProxies  []Proxy
	OwnedBackends []Backend
	Tokens        []Token
}

var DB *gorm.DB

func InitializeDatabaseDialector() (gorm.Dialector, error) {
	databaseBackend := os.Getenv("HERMES_DATABASE_BACKEND")

	switch databaseBackend {
	case "sqlite":
		filePath := os.Getenv("HERMES_SQLITE_FILEPATH")

		if filePath == "" {
			return nil, fmt.Errorf("sqlite database file not specified (missing HERMES_SQLITE_FILEPATH)")
		}

		return sqlite.Open(filePath), nil
	case "":
		return nil, fmt.Errorf("no database backend specified in environment variables (missing HERMES_DATABASE_BACKEND)")
	default:
		return nil, fmt.Errorf("unknown database backend specified: %s", os.Getenv(databaseBackend))
	}
}

func InitializeDatabase(config *gorm.Config) error {
	var err error

	dialector, err := InitializeDatabaseDialector()

	if err != nil {
		return fmt.Errorf("failed to initialize physical database: %s", err)
	}

	DB, err = gorm.Open(dialector, config)

	if err != nil {
		return fmt.Errorf("failed to open database: %s", err)
	}

	return nil
}

func DoDatabaseMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&Proxy{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Backend{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Permission{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&Token{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}

	return nil
}
