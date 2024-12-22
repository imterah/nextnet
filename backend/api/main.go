package main

import (
	"os"
	"strings"

	"git.terah.dev/imterah/hermes/api/controllers/v1/users"
	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	logLevel := os.Getenv("HERMES_LOG_LEVEL")
	developmentMode := false

	if os.Getenv("HERMES_DEVELOPMENT_MODE") != "" {
		developmentMode = true
	}

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

	log.Info("Hermes is initializing...")
	log.Debug("Initializing database and opening it...")

	err := dbcore.InitializeDatabase(&gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}

	log.Debug("Running database migrations...")

	if err := dbcore.DoDatabaseMigrations(dbcore.DB); err != nil {
		log.Fatalf("Failed to run database migrations: %s", err)
	}

	log.Debug("Initializing the JWT subsystem...")

	if err := jwtcore.SetupJWT(); err != nil {
		log.Fatalf("Failed to initialize the JWT subsystem: %s", err.Error())
	}

	log.Debug("Initializing API...")

	if !developmentMode {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	listeningAddress := os.Getenv("HERMES_LISTENING_ADDRESS")

	if listeningAddress == "" {
		if developmentMode {
			listeningAddress = "localhost:8000"
		} else {
			listeningAddress = "0.0.0.0:8000"
		}
	}

	trustedProxiesString := os.Getenv("HERMES_TRUSTED_HTTP_PROXIES")

	if trustedProxiesString != "" {
		trustedProxies := strings.Split(trustedProxiesString, ",")

		engine.ForwardedByClientIP = true
		engine.SetTrustedProxies(trustedProxies)
	} else {
		engine.ForwardedByClientIP = false
		engine.SetTrustedProxies(nil)
	}

	// Initialize routes
	engine.POST("/api/v1/users/create", users.CreateUser)
	engine.POST("/api/v1/users/login", users.LoginUser)

	log.Infof("Listening on: %s", listeningAddress)
	err = engine.Run(listeningAddress)

	if err != nil {
		log.Fatalf("Error running web server: %s", err.Error())
	}
}
