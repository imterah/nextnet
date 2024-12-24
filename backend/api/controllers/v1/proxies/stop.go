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

type ProxyStopRequest struct {
	Token string `validate:"required" json:"token"`
	ID    uint   `validate:"required" json:"id"`
}

type ProxyStopResponse struct {
	Success bool `json:"success"`
}

func StopProxy(c *gin.Context) {
	var req ProxyStopRequest

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

	if !permissions.UserHasPermission(user, "routes.stop") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	var proxy *dbcore.Proxy
	proxyRequest := dbcore.DB.Where("id = ?", req.ID).Find(&proxy)

	if proxyRequest.Error != nil {
		log.Warnf("failed to find if proxy exists or not: %s", proxyRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if forward rule exists",
		})

		return
	}

	proxyExists := proxyRequest.RowsAffected > 0

	if !proxyExists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Forward rule doesn't exist",
		})

		return
	}

	backend := backendruntime.RunningBackends[proxy.BackendID]

	backend.RuntimeCommands <- commonbackend.RemoveProxy{
		Type:       "removeProxy",
		SourceIP:   proxy.SourceIP,
		SourcePort: proxy.SourcePort,
		DestPort:   proxy.DestinationPort,
		Protocol:   proxy.Protocol,
	}

	c.JSON(http.StatusOK, &ProxyStopResponse{
		Success: true,
	})
}
