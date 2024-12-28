package backends

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"git.terah.dev/imterah/hermes/backend/api/backendruntime"
	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"git.terah.dev/imterah/hermes/backend/api/jwtcore"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BackendLookupRequest struct {
	Token       string `validate:"required"`
	BackendID   *uint  `json:"id"`
	Name        *string
	Description *string
	Backend     *string
}

type SanitizedBackend struct {
	Name              string   `json:"name"`
	BackendID         uint     `json:"id"`
	OwnerID           uint     `json:"ownerID"`
	Description       *string  `json:"description,omitempty"`
	Backend           string   `json:"backend"`
	BackendParameters *string  `json:"connectionDetails,omitempty"`
	Logs              []string `json:"logs"`
}

type LookupResponse struct {
	Success bool                `json:"success"`
	Data    []*SanitizedBackend `json:"data"`
}

func LookupBackend(c *gin.Context) {
	var req BackendLookupRequest

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

	if !permissions.UserHasPermission(user, "backends.visible") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	backends := []dbcore.Backend{}
	queryString := []string{}
	queryParameters := []interface{}{}

	if req.BackendID != nil {
		queryString = append(queryString, "id = ?")
		queryParameters = append(queryParameters, req.BackendID)
	}

	if req.Name != nil {
		queryString = append(queryString, "name = ?")
		queryParameters = append(queryParameters, req.Name)
	}

	if req.Description != nil {
		queryString = append(queryString, "description = ?")
		queryParameters = append(queryParameters, req.Description)
	}

	if req.Backend != nil {
		queryString = append(queryString, "is_bot = ?")
		queryParameters = append(queryParameters, req.Backend)
	}

	if err := dbcore.DB.Where(strings.Join(queryString, " AND "), queryParameters...).Find(&backends).Error; err != nil {
		log.Warnf("Failed to get backends: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get backends",
		})

		return
	}

	sanitizedBackends := make([]*SanitizedBackend, len(backends))
	hasSecretVisibility := permissions.UserHasPermission(user, "backends.secretVis")

	for backendIndex, backend := range backends {
		foundBackend, ok := backendruntime.RunningBackends[backend.ID]

		if !ok {
			log.Warnf("Failed to get backend #%d controller", backend.ID)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get backends",
			})

			return
		}

		sanitizedBackends[backendIndex] = &SanitizedBackend{
			BackendID:   backend.ID,
			OwnerID:     backend.UserID,
			Name:        backend.Name,
			Description: backend.Description,
			Backend:     backend.Backend,
			Logs:        foundBackend.Logs,
		}

		if backend.UserID == user.ID || hasSecretVisibility {
			backendParametersBytes, err := base64.StdEncoding.DecodeString(backend.BackendParameters)

			if err != nil {
				log.Warnf("Failed to decode base64 backend parameters: %s", err.Error())
			}

			backendParameters := string(backendParametersBytes)
			sanitizedBackends[backendIndex].BackendParameters = &backendParameters
		}
	}

	c.JSON(http.StatusOK, &LookupResponse{
		Success: true,
		Data:    sanitizedBackends,
	})
}
