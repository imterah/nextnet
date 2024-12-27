package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginRequest struct {
	Username *string
	Email    *string

	Password string `validate:"required"`
}

func LoginUser(c *gin.Context) {
	var req UserLoginRequest

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

	if req.Email == nil && req.Username == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing both email and username in body",
		})

		return
	}

	userFindRequestArguments := make([]interface{}, 1)
	userFindRequest := ""

	if req.Email != nil {
		userFindRequestArguments[0] = &req.Email
		userFindRequest += "email = ?"
	}

	if req.Username != nil {
		userFindRequestArguments[0] = &req.Username
		userFindRequest += "username = ?"
	}

	var user *dbcore.User
	userRequest := dbcore.DB.Where(userFindRequest, userFindRequestArguments...).Find(&user)

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
			"error": "User not found",
		})

		return
	}

	decodedPassword := make([]byte, base64.StdEncoding.DecodedLen(len(user.Password)))
	_, err := base64.StdEncoding.Decode(decodedPassword, []byte(user.Password))

	if err != nil {
		log.Warnf("failed to decode password in database: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse database result for password",
		})

		return
	}

	err = bcrypt.CompareHashAndPassword(decodedPassword, []byte(req.Password))

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Invalid password",
		})

		return
	}

	tokenRandomData := make([]byte, 80)

	if _, err := rand.Read(tokenRandomData); err != nil {
		log.Warnf("Failed to read random data to use as token: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})

		return
	}

	token := &dbcore.Token{
		UserID: user.ID,

		Token:          base64.StdEncoding.EncodeToString(tokenRandomData),
		DisableExpiry:  forceNoExpiryTokens,
		CreationIPAddr: c.ClientIP(),
	}

	if result := dbcore.DB.Create(&token); result.Error != nil {
		log.Warnf("Failed to create user: %s", result.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add refresh token into database",
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
		"success":      true,
		"token":        jwt,
		"refreshToken": base64.StdEncoding.EncodeToString(tokenRandomData),
	})
}
