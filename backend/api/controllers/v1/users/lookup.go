package users

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

type UserLookupRequest struct {
	Token    string  `validate:"required"`
	UID      *uint   `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	IsBot    *bool   `json:"isServiceAccount"`
}

type SanitizedUsers struct {
	UID      uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsBot    bool   `json:"isServiceAccount"`
}

type LookupResponse struct {
	Success bool              `json:"success"`
	Data    []*SanitizedUsers `json:"data"`
}

func LookupUser(c *gin.Context) {
	var req UserLookupRequest

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

	users := []dbcore.User{}
	queryString := []string{}
	queryParameters := []interface{}{}

	if !permissions.UserHasPermission(user, "users.lookup") {
		queryString = append(queryString, "id = ?")
		queryParameters = append(queryParameters, user.ID)
	} else if permissions.UserHasPermission(user, "users.lookup") && req.UID != nil {
		queryString = append(queryString, "id = ?")
		queryParameters = append(queryParameters, req.UID)
	}

	if req.Name != nil {
		queryString = append(queryString, "name = ?")
		queryParameters = append(queryParameters, req.Name)
	}

	if req.Email != nil {
		queryString = append(queryString, "email = ?")
		queryParameters = append(queryParameters, req.Email)
	}

	if req.IsBot != nil {
		queryString = append(queryString, "is_bot = ?")
		queryParameters = append(queryParameters, req.IsBot)
	}

	if err := dbcore.DB.Where(strings.Join(queryString, " AND "), queryParameters...).Find(&users).Error; err != nil {
		log.Warnf("Failed to get users: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get users",
		})

		return
	}

	sanitizedUsers := make([]*SanitizedUsers, len(users))

	for userIndex, user := range users {
		isBot := false

		if user.IsBot != nil {
			isBot = *user.IsBot
		}

		sanitizedUsers[userIndex] = &SanitizedUsers{
			UID:      user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Username: user.Username,
			IsBot:    isBot,
		}
	}

	c.JSON(http.StatusOK, &LookupResponse{
		Success: true,
		Data:    sanitizedUsers,
	})
}
