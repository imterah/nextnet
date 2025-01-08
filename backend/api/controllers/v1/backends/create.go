package backends

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/backend/api/backendruntime"
	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"git.terah.dev/imterah/hermes/backend/api/jwtcore"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BackendCreationRequest struct {
	Token             string `validate:"required"`
	Name              string `validate:"required"`
	Description       *string
	Backend           string      `validate:"required"`
	BackendParameters interface{} `json:"connectionDetails" validate:"required"`
}

func CreateBackend(c *gin.Context) {
	var req BackendCreationRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to parse body: %s", err.Error()),
		})

		return
	}

	if err := validator.New().Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to validate body: %s", err.Error()),
		})

		return
	}

	user, err := jwtcore.GetUserFromJWT(req.Token)

	if err != nil {
		if err.Error() == "token is expired" || err.Error() == "user does not exist" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})

			return
		} else {
			log.Warnf("Failed to get user from the provided JWT token: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse token",
			})

			return
		}
	}

	if !permissions.UserHasPermission(user, "backends.add") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	var backendParameters []byte

	switch parameters := req.BackendParameters.(type) {
	case string:
		backendParameters = []byte(parameters)
	case map[string]interface{}:
		backendParameters, err = json.Marshal(parameters)

		if err != nil {
			log.Warnf("Failed to marshal JSON recieved as BackendParameters: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to prepare parameters",
			})

			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid type for connectionDetails (recieved %T)", parameters),
		})

		return
	}

	var backendRuntimeFilePath string

	for _, runtime := range backendruntime.AvailableBackends {
		if runtime.Name == req.Backend {
			backendRuntimeFilePath = runtime.Path
		}
	}

	if backendRuntimeFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported backend recieved",
		})

		return
	}

	backend := backendruntime.NewBackend(backendRuntimeFilePath)
	err = backend.Start()

	if err != nil {
		log.Warnf("Failed to start backend: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start backend",
		})

		return
	}

	backendParamCheckResponse, err := backend.ProcessCommand(&commonbackend.CheckServerParameters{
		Type:      "checkServerParameters",
		Arguments: backendParameters,
	})

	if err != nil {
		log.Warnf("Failed to get response for backend: %s", err.Error())

		err = backend.Stop()

		if err != nil {
			log.Warnf("Failed to stop backend: %s", err.Error())
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get status response from backend",
		})

		return
	}

	switch responseMessage := backendParamCheckResponse.(type) {
	case *commonbackend.CheckParametersResponse:
		if responseMessage.InResponseTo != "checkServerParameters" {
			log.Errorf("Got illegal response to CheckServerParameters: %s", responseMessage.InResponseTo)

			err = backend.Stop()

			if err != nil {
				log.Warnf("Failed to stop backend: %s", err.Error())
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get status response from backend",
			})

			return
		}

		if !responseMessage.IsValid {
			err = backend.Stop()

			if err != nil {
				log.Warnf("Failed to stop backend: %s", err.Error())
			}

			var errorMessage string

			if responseMessage.Message == "" {
				errorMessage = "Unkown error while trying to parse connectionDetails"
			} else {
				errorMessage = fmt.Sprintf("Invalid backend parameters: %s", responseMessage.Message)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": errorMessage,
			})

			return
		}
	default:
		log.Warnf("Got illegal response type for backend: %T", responseMessage)
	}

	log.Info("Passed backend checks successfully")

	backendInDatabase := &dbcore.Backend{
		UserID:            user.ID,
		Name:              req.Name,
		Description:       req.Description,
		Backend:           req.Backend,
		BackendParameters: base64.StdEncoding.EncodeToString(backendParameters),
	}

	if result := dbcore.DB.Create(&backendInDatabase); result.Error != nil {
		log.Warnf("Failed to create backend: %s", result.Error.Error())

		err = backend.Stop()

		if err != nil {
			log.Warnf("Failed to stop backend: %s", err.Error())
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add backend into database",
		})

		return
	}

	backendStartResponse, err := backend.ProcessCommand(&commonbackend.Start{
		Type:      "start",
		Arguments: backendParameters,
	})

	if err != nil {
		log.Warnf("Failed to get response for backend: %s", err.Error())

		err = backend.Stop()

		if err != nil {
			log.Warnf("Failed to stop backend: %s", err.Error())
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get status response from backend",
		})

		return
	}

	switch responseMessage := backendStartResponse.(type) {
	case *commonbackend.BackendStatusResponse:
		if !responseMessage.IsRunning {
			err = backend.Stop()

			if err != nil {
				log.Warnf("Failed to start backend: %s", err.Error())
			}

			var errorMessage string

			if responseMessage.Message == "" {
				errorMessage = "Unkown error while trying to start the backend"
			} else {
				errorMessage = fmt.Sprintf("Failed to start backend: %s", responseMessage.Message)
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": errorMessage,
			})

			return
		}
	default:
		log.Warnf("Got illegal response type for backend: %T", responseMessage)
	}

	backendruntime.RunningBackends[backendInDatabase.ID] = backend

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
