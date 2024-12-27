package users

import (
	"fmt"
	"net/http"
	"time"

	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserRefreshRequest struct {
	Token string `validate:"required"`
}

func RefreshUserToken(c *gin.Context) {
	var req UserRefreshRequest

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

	var tokenInDatabase *dbcore.Token
	tokenRequest := dbcore.DB.Where("token = ?", req.Token).Find(&tokenInDatabase)

	if tokenRequest.Error != nil {
		log.Warnf("failed to find if token exists or not: %s", tokenRequest.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if token exists",
		})

		return
	}

	tokenExists := tokenRequest.RowsAffected > 0

	if !tokenExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token not found",
		})

		return
	}

	// First, we check to make sure that the key expiry is disabled before checking if the key is expired.
	// Then, we check if the IP addresses differ, or if it has been 7 days since the token has been created.
	if !tokenInDatabase.DisableExpiry && (c.ClientIP() != tokenInDatabase.CreationIPAddr || time.Now().Before(tokenInDatabase.CreatedAt.Add((24*7)*time.Hour))) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Token has expired",
		})

		tx := dbcore.DB.Delete(tokenInDatabase)

		if tx.Error != nil {
			log.Warnf("Failed to delete expired token from database: %s", tx.Error.Error())
		}

		return
	}

	// Get the user to check if the user exists before doing anything
	var user *dbcore.User
	userRequest := dbcore.DB.Where("id = ?", tokenInDatabase.UserID).Find(&user)

	if tokenRequest.Error != nil {
		log.Warnf("failed to find if token user or not: %s", userRequest.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find user",
		})

		return
	}

	userExists := userRequest.RowsAffected > 0

	if !userExists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User not found",
		})

		return
	}

	jwt, err := jwtcore.Generate(user.ID)

	if err != nil {
		log.Warnf("Failed to generate JWT: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   jwt,
	})
}
