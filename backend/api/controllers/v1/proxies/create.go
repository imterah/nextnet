package proxies

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"git.terah.dev/imterah/hermes/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProxyCreationRequest struct {
	Token           string  `validate:"required" json:"token"`
	Name            string  `validate:"required" json:"name"`
	Description     *string `json:"description"`
	Protcol         string  `validate:"required" json:"protcol"`
	SourceIP        string  `validate:"required" json:"source_ip"`
	SourcePort      uint16  `validate:"required" json:"source_port"`
	DestinationPort uint16  `validate:"required" json:"destination_port"`
	ProviderID      uint    `validate:"required" json:"provider_id"`
	AutoStart       bool    `json:"auto_start"`
}

type ProxyCreationResponse struct {
	Success bool `json:"success"`
	Id      uint `json:"id"`
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

	if req.Protcol != "tcp" && req.Protcol != "udp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Body protocol must be either 'tcp' or 'udp'",
		})

		return
	}

	var backend dbcore.Backend
	backendRequest := dbcore.DB.Where("id = ?", req.ProviderID).First(&backend)
	if backendRequest.Error != nil {
		log.Warnf("failed to find if backend exists or not: %s", backendRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if provider exists",
		})
	}

	backendExists := backendRequest.RowsAffected > 0
	if !backendExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not find provider",
		})
	}

	proxy := &dbcore.Proxy{
		UserID:          user.ID,
		BackendID:       req.ProviderID,
		Name:            req.Name,
		Description:     req.Description,
		Protocol:        req.Protcol,
		SourceIP:        req.SourceIP,
		SourcePort:      req.SourcePort,
		DestinationPort: req.DestinationPort,
		AutoStart:       req.AutoStart,
	}

	if result := dbcore.DB.Create(proxy); result.Error != nil {
		log.Warnf("failed to create proxy: %s", result.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add forward rule to database",
		})
	}

	c.JSON(http.StatusOK, &ProxyCreationResponse{
		Success: true,
		Id:      proxy.ID,
	})
}
