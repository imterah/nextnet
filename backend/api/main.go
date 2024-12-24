package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.terah.dev/imterah/hermes/api/backendruntime"
	"git.terah.dev/imterah/hermes/api/controllers/v1/backends"
	"git.terah.dev/imterah/hermes/api/controllers/v1/proxies"
	"git.terah.dev/imterah/hermes/api/controllers/v1/users"
	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"git.terah.dev/imterah/hermes/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func entrypoint(cCtx *cli.Context) error {
	developmentMode := false

	if os.Getenv("HERMES_DEVELOPMENT_MODE") != "" {
		log.Warn("You have development mode enabled. This may weaken security.")
		developmentMode = true
	}

	log.Info("Hermes is initializing...")
	log.Debug("Initializing database and opening it...")

	err := dbcore.InitializeDatabase(&gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}

	log.Debug("Running database migrations...")

	if err := dbcore.DoDatabaseMigrations(dbcore.DB); err != nil {
		return fmt.Errorf("Failed to run database migrations: %s", err)
	}

	log.Debug("Initializing the JWT subsystem...")

	if err := jwtcore.SetupJWT(); err != nil {
		return fmt.Errorf("Failed to initialize the JWT subsystem: %s", err.Error())
	}

	log.Debug("Initializing the backend subsystem...")

	backendMetadataPath := cCtx.String("backends-path")
	backendMetadata, err := os.ReadFile(backendMetadataPath)

	if err != nil {
		return fmt.Errorf("Failed to read backends: %s", err.Error())
	}

	availableBackends := []*backendruntime.Backend{}
	err = json.Unmarshal(backendMetadata, &availableBackends)

	if err != nil {
		return fmt.Errorf("Failed to parse backends: %s", err.Error())
	}

	for _, backend := range availableBackends {
		backend.Path = path.Join(filepath.Dir(backendMetadataPath), backend.Path)
	}

	backendruntime.Init(availableBackends)

	log.Debug("Enumerating backends...")

	backendList := []dbcore.Backend{}

	if err := dbcore.DB.Find(&backendList).Error; err != nil {
		return fmt.Errorf("Failed to enumerate backends: %s", err.Error())
	}

	for _, backend := range backendList {
		log.Infof("Starting up backend #%d: %s", backend.ID, backend.Name)

		var backendRuntimeFilePath string

		for _, runtime := range backendruntime.AvailableBackends {
			if runtime.Name == backend.Backend {
				backendRuntimeFilePath = runtime.Path
			}
		}

		if backendRuntimeFilePath == "" {
			log.Errorf("Unsupported backend recieved for ID %d: %s", backend.ID, backend.Backend)
			continue
		}

		backendInstance := backendruntime.NewBackend(backendRuntimeFilePath)
		err = backendInstance.Start()

		if err != nil {
			log.Errorf("Failed to start backend #%d: %s", backend.ID, err.Error())
			continue
		}

		backendParameters, err := base64.StdEncoding.DecodeString(backend.BackendParameters)

		if err != nil {
			log.Errorf("Failed to decode backend parameters for backend #%d: %s", backend.ID, err.Error())
			continue
		}

		backendInstance.RuntimeCommands <- &commonbackend.Start{
			Type:      "start",
			Arguments: backendParameters,
		}

		backendStartResponse := <-backendInstance.RuntimeCommands

		switch responseMessage := backendStartResponse.(type) {
		case error:
			log.Warnf("Failed to get response for backend #%d: %s", backend.ID, responseMessage.Error())

			err = backendInstance.Stop()

			if err != nil {
				log.Warnf("Failed to stop backend: %s", err.Error())
			}

			continue
		case *commonbackend.BackendStatusResponse:
			if !responseMessage.IsRunning {
				err = backendInstance.Stop()

				if err != nil {
					log.Warnf("Failed to start backend: %s", err.Error())
				}

				if responseMessage.Message == "" {
					log.Errorf("Unkown error while trying to start the backend #%d", backend.ID)
				} else {
					log.Errorf("Failed to start backend: %s", responseMessage.Message)
				}

				continue
			}
		default:
			log.Errorf("Got illegal response type for backend #%d: %T", backend.ID, responseMessage)
			continue
		}

		backendruntime.RunningBackends[backend.ID] = backendInstance
		log.Infof("Successfully started backend #%d", backend.ID)
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
	engine.POST("/api/v1/users/refresh", users.RefreshUserToken)
	engine.POST("/api/v1/users/remove", users.RemoveUser)
	engine.POST("/api/v1/users/lookup", users.LookupUser)

	engine.POST("/api/v1/backends/create", backends.CreateBackend)
	engine.POST("/api/v1/backends/remove", backends.RemoveBackend)
	engine.POST("/api/v1/backends/lookup", backends.LookupBackend)

	engine.POST("/api/v1/forward/create", proxies.CreateProxy)
	engine.POST("/api/v1/forward/lookup", proxies.LookupProxy)
	engine.POST("/api/v1/forward/remove", proxies.RemoveProxy)
	engine.POST("/api/v1/forward/start", proxies.StartProxy)
	engine.POST("/api/v1/forward/stop", proxies.StopProxy)

	log.Infof("Listening on '%s'", listeningAddress)
	err = engine.Run(listeningAddress)

	if err != nil {
		return fmt.Errorf("Error running web server: %s", err.Error())
	}

	return nil
}

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

	app := &cli.App{
		Name:  "hermes",
		Usage: "port forwarding across boundaries",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "backends-path",
				Aliases:  []string{"b"},
				Usage:    "path to the backend manifest file",
				Required: true,
			},
		},
		Action: entrypoint,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
