package proxies

import (
	"fmt"
	"net/http"
	"strings"

	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"git.terah.dev/imterah/hermes/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProxyLookupRequest struct {
	Token           string  `validate:"required" json:"token"`
	Id              *uint   `json:"id"`
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Protocol        *string `json:"protocol"`
	SourceIP        *string `json:"source_ip"`
	SourcePort      *uint16 `json:"source_port"`
	DestinationPort *uint16 `json:"destination_port"`
	ProviderID      *uint   `json:"provider_id"`
	AutoStart       *bool   `json:"auto_start"`
}

type SanitizedProxy struct {
	Id              uint    `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	Protcol         string  `json:"protcol"`
	SourceIP        string  `json:"source_ip"`
	SourcePort      uint16  `json:"source_port"`
	DestinationPort uint16  `json:"destination_port"`
	ProviderID      uint    `json:"provider_id"`
	AutoStart       bool    `json:"auto_start"`
}

type ProxyLookupResponse struct {
	Success bool              `json:"success"`
	Data    []*SanitizedProxy `json:"data"`
}

func LookupProxy(c *gin.Context) {
	var req ProxyLookupRequest

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

	if !permissions.UserHasPermission(user, "routes.visible") {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Missing permissions",
		})

		return
	}

	if *req.Protcol != "tcp" && *req.Protcol != "udp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Protocol specified in body must either be 'tcp' or 'udp'",
		})
	}

	proxies := []dbcore.Proxy{}
	queryString := []string{}
	queryParameters := []interface{}{}

	if req.Id != nil {
		queryString = append(queryString, "id = ?")
		queryParameters = append(queryParameters, req.Id)
	}
	if req.Name != nil {
		queryString = append(queryString, "name = ?")
		queryParameters = append(queryParameters, req.Name)
	}
	if req.Description != nil {
		queryString = append(queryString, "description = ?")
		queryParameters = append(queryParameters, req.Description)
	}
	if req.SourceIP != nil {
		queryString = append(queryString, "name = ?")
		queryParameters = append(queryParameters, req.Name)
	}
	if req.SourcePort != nil {
		queryString = append(queryString, "sourceport = ?")
		queryParameters = append(queryParameters, req.SourcePort)
	}
	if req.DestinationPort != nil {
		queryString = append(queryString, "destinationport = ?")
		queryParameters = append(queryParameters, req.DestinationPort)
	}
	if req.ProviderID != nil {
		queryString = append(queryString, "backendid = ?")
		queryParameters = append(queryParameters, req.ProviderID)
	}
	if req.AutoStart != nil {
		queryString = append(queryString, "autostart = ?")
		queryParameters = append(queryParameters, req.AutoStart)
	}
	if req.Protocol != nil {
		queryString = append(queryString, "protocol = ?")
		queryParameters = append(queryParameters, req.Protocol)
	}

	if err := dbcore.DB.Where(strings.Join(queryString, " AND "), queryParameters...).Find(&proxies).Error; err != nil {
		log.Warnf("failed to get proxies: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get forward rules",
		})

		return
	}

	sanitizedProxies := make([]*SanitizedProxy, len(proxies))

	for proxyIndex, proxy := range proxies {
		description := ""
		if proxy.Description != nil {
			description = *proxy.Description
		}

		sanitizedProxies[proxyIndex] = &SanitizedProxy{
			Id:              proxy.ID,
			Name:            proxy.Name,
			Description:     &description,
			Protcol:         proxy.Protocol,
			SourceIP:        proxy.SourceIP,
			SourcePort:      proxy.SourcePort,
			DestinationPort: proxy.DestinationPort,
			ProviderID:      proxy.BackendID,
			AutoStart:       proxy.AutoStart,
		}
	}

	c.JSON(http.StatusOK, &ProxyLookupResponse{
		Success: true,
		Data:    sanitizedProxies,
	})
}
