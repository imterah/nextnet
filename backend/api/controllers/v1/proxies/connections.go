package proxies

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/api/backendruntime"
	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"git.terah.dev/imterah/hermes/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ConnectionsRequest struct {
	Token string `validate:"required" json:"token"`
	Id    uint   `validate:"required" json:"id"`
}

type SanitizedBackends struct {
	UserID            uint   `json:"user_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Backend           string `json:"backend"`
	BackendParameters string `json:"backend_parameters"`
}

type ConnectionsResponse struct {
	Success bool                 `json:"success"`
	Data    []*SanitizedBackends `json:"data"`
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
				"error": "failed to parse token",
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

	var routes []dbcore.Proxy
	routesRequest := dbcore.DB.Where("id = ?", req.Id).First(&routes)

	if routesRequest.Error != nil {
		log.Warnf("failed to find proxy: %s", routesRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find forward entry",
		})

		return
	}

	routesExist := routesRequest.RowsAffected > 0

	if !routesExist {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No forward entry found",
		})

		return
	}

	// FIXME(greysoh): not finished
	var backends []dbcore.Backend

	sanitizedBackends := make([]*SanitizedBackends, len(backends))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sanitizedBackends,
	})
}
