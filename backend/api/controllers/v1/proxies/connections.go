package proxies

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/api/backendruntime"
	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"git.terah.dev/imterah/hermes/api/permissions"
	"git.terah.dev/imterah/hermes/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ConnectionsRequest struct {
	Token string `validate:"required" json:"token"`
	Id    uint   `validate:"required" json:"id"`
}

type ConnectionDetailsForConnection struct {
	SourceIP   string `json:"sourceIP"`
	SourcePort uint16 `json:"sourcePort"`
	DestPort   uint16 `json:"destPort"`
}

type SanitizedConnection struct {
	ClientIP string `json:"ip"`
	Port     uint16 `json:"port"`

	ConnectionDetails *ConnectionDetailsForConnection `json:"connectionDetails"`
}

type ConnectionsResponse struct {
	Success bool                   `json:"success"`
	Data    []*SanitizedConnection `json:"data"`
}

func Connections(c *gin.Context) {
	var req ConnectionsRequest

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

	if !permissions.UserHasPermission(user, "routes.visibleConn") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	var proxy dbcore.Proxy
	proxyRequest := dbcore.DB.Where("id = ?", req.Id).First(&proxy)

	if proxyRequest.Error != nil {
		log.Warnf("failed to find proxy: %s", proxyRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find forward entry",
		})

		return
	}

	proxyExists := proxyRequest.RowsAffected > 0

	if !proxyExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No forward entry found",
		})

		return
	}

	var backend dbcore.Backend
	backendRequest := dbcore.DB.Where("id = ?", proxy.BackendID).First(&backend)

	if backendRequest.Error != nil {
		log.Warnf("failed to find backend: %s", backendRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find backend entry",
		})

		return
	}

	backendExists := backendRequest.RowsAffected > 0

	if !backendExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No forward entry found",
		})

		return
	}

	backendRuntime, ok := backendruntime.RunningBackends[backend.ID]

	if !ok {
		log.Warnf("Couldn't fetch backend runtime from backend ID #%d", backend.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't fetch backend runtime",
		})

		return
	}

	backendRuntime.RuntimeCommands <- &commonbackend.ProxyConnectionsRequest{
		Type: "proxyConnectionsRequest",
	}

	backendResponse := <-backendRuntime.RuntimeCommands

	switch responseMessage := backendResponse.(type) {
	case error:
		log.Warnf("Failed to get response for backend: %s", responseMessage.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get status response from backend",
		})
	case *commonbackend.ProxyConnectionsResponse:
		sanitizedConnections := []*SanitizedConnection{}

		for _, connection := range responseMessage.Connections {
			if connection.SourceIP == proxy.SourceIP && connection.SourcePort == proxy.SourcePort && proxy.DestinationPort == proxy.DestinationPort {
				sanitizedConnections = append(sanitizedConnections, &SanitizedConnection{
					ClientIP: connection.ClientIP,
					Port:     connection.ClientPort,

					ConnectionDetails: &ConnectionDetailsForConnection{
						SourceIP:   proxy.SourceIP,
						SourcePort: proxy.SourcePort,
						DestPort:   proxy.DestinationPort,
					},
				})
			}
		}

		c.JSON(http.StatusOK, &ConnectionsResponse{
			Success: true,
			Data:    sanitizedConnections,
		})
	default:
		log.Warnf("Got illegal response type for backend: %T", responseMessage)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Got illegal response type",
		})
	}
}
