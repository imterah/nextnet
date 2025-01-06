package proxies

import (
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

type ProxyCreationRequest struct {
	Token           string  `validate:"required" json:"token"`
	Name            string  `validate:"required" json:"name"`
	Description     *string `json:"description"`
	Protocol        string  `validate:"required" json:"protocol"`
	SourceIP        string  `validate:"required" json:"sourceIP"`
	SourcePort      uint16  `validate:"required" json:"sourcePort"`
	DestinationPort uint16  `validate:"required" json:"destinationPort"`
	ProviderID      uint    `validate:"required" json:"providerID"`
	AutoStart       *bool   `json:"autoStart"`
}

func CreateProxy(c *gin.Context) {
	var req ProxyCreationRequest

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

	if !permissions.UserHasPermission(user, "routes.add") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	if req.Protocol != "tcp" && req.Protocol != "udp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Protocol must be either 'tcp' or 'udp'",
		})

		return
	}

	var backend dbcore.Backend
	backendRequest := dbcore.DB.Where("id = ?", req.ProviderID).First(&backend)

	if backendRequest.Error != nil {
		log.Warnf("failed to find if backend exists or not: %s", backendRequest.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if backend exists",
		})
	}

	backendExists := backendRequest.RowsAffected > 0

	if !backendExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not find backend",
		})
	}

	autoStart := false

	if req.AutoStart != nil {
		autoStart = *req.AutoStart
	}

	proxy := &dbcore.Proxy{
		UserID:          user.ID,
		BackendID:       req.ProviderID,
		Name:            req.Name,
		Description:     req.Description,
		Protocol:        req.Protocol,
		SourceIP:        req.SourceIP,
		SourcePort:      req.SourcePort,
		DestinationPort: req.DestinationPort,
		AutoStart:       autoStart,
	}

	if result := dbcore.DB.Create(proxy); result.Error != nil {
		log.Warnf("failed to create proxy: %s", result.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add forward rule to database",
		})
	}

	if autoStart {
		backend, ok := backendruntime.RunningBackends[proxy.BackendID]

		if !ok {
			log.Warnf("Couldn't fetch backend runtime from backend ID #%d", proxy.BackendID)

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"id":      proxy.ID,
			})

			return
		}

		backendResponse, err := backend.ProcessCommand(&commonbackend.AddProxy{
			Type:       "addProxy",
			SourceIP:   proxy.SourceIP,
			SourcePort: proxy.SourcePort,
			DestPort:   proxy.DestinationPort,
			Protocol:   proxy.Protocol,
		})

		if err != nil {
			log.Warnf("Failed to get response for backend #%d: %s", proxy.BackendID, err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get response from backend",
			})

			return
		}

		switch responseMessage := backendResponse.(type) {
		case *commonbackend.ProxyStatusResponse:
			if !responseMessage.IsActive {
				log.Warnf("Failed to start proxy for backend #%d", proxy.BackendID)
			}
		default:
			log.Errorf("Got illegal response type for backend #%d: %T", proxy.BackendID, responseMessage)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"id":      proxy.ID,
	})
}
