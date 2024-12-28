package users

import (
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"git.terah.dev/imterah/hermes/backend/api/jwtcore"
	"git.terah.dev/imterah/hermes/backend/api/permissions"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserRemovalRequest struct {
	Token string `validate:"required"`
	UID   *uint  `json:"uid"`
}

func RemoveUser(c *gin.Context) {
	var req UserRemovalRequest

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

	uid := user.ID

	if req.UID != nil {
		uid = *req.UID

		if uid != user.ID && !permissions.UserHasPermission(user, "users.remove") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Missing permissions",
			})

			return
		}
	}

	// Make sure the user exists first if we have a custom UserID

	if uid != user.ID {
		var customUser *dbcore.User
		userRequest := dbcore.DB.Where("id = ?", uid).Find(customUser)

		if userRequest.Error != nil {
			log.Warnf("failed to find if user exists or not: %s", userRequest.Error.Error())

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to find if user exists",
			})

			return
		}

		userExists := userRequest.RowsAffected > 0

		if !userExists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User doesn't exist",
			})

			return
		}
	}

	dbcore.DB.Select("Tokens", "Permissions", "Proxys", "Backends").Where("id = ?", uid).Delete(user)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
