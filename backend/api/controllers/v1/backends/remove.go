package backends

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/backend/api/backendruntime"
	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"git.terah.dev/imterah/hermes/backend/api/jwtcore"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BackendRemovalRequest struct {
	Token     string `validate:"required"`
	BackendID uint   `json:"id" validate:"required"`
}

func RemoveBackend(c *gin.Context) {
	var req BackendRemovalRequest

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

	if !permissions.UserHasPermission(user, "backends.remove") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	var backend *dbcore.Backend
	backendRequest := dbcore.DB.Where("id = ?", req.BackendID).Find(&backend)

	if backendRequest.Error != nil {
		log.Warnf("failed to find if backend exists or not: %s", backendRequest.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if backend exists",
		})

		return
	}

	backendExists := backendRequest.RowsAffected > 0

	if !backendExists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Backend doesn't exist",
		})

		return
	}

	if err := dbcore.DB.Delete(backend).Error; err != nil {
		log.Warnf("failed to delete backend: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete backend",
		})

		return
	}

	backendInstance, ok := backendruntime.RunningBackends[req.BackendID]

	if ok {
		err = backendInstance.Stop()

		if err != nil {
			log.Warnf("Failed to stop backend: %s", err.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Backend deleted, but failed to stop",
			})

			delete(backendruntime.RunningBackends, req.BackendID)
			return
		}

		delete(backendruntime.RunningBackends, req.BackendID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
