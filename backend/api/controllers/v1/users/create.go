package users

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"

	"git.terah.dev/imterah/hermes/api/constants"
	"git.terah.dev/imterah/hermes/api/dbcore"
	"git.terah.dev/imterah/hermes/api/jwtcore"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserCreationRequest struct {
	Name     string `validate:"required"`
	Email    string `validate:"required"`
	Password string `validate:"required"`
	Username string `validate:"required"`

	// TODO: implement support
	ExistingUserToken string `json:"token"`
	IsBot             bool
}

var (
	signupEnabled       bool
	unsafeSignup        bool
	forceNoExpiryTokens bool
)

func init() {
	signupEnabled = os.Getenv("HERMES_SIGNUP_ENABLED") != ""
	unsafeSignup = os.Getenv("HERMES_UNSAFE_ADMIN_SIGNUP_ENABLED") != ""
	forceNoExpiryTokens = os.Getenv("HERMES_FORCE_DISABLE_REFRESH_TOKEN_EXPIRY") != ""
}

func CreateUser(c *gin.Context) {
	var req UserCreationRequest

	if !signupEnabled && !unsafeSignup {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Signing up is not enabled at this time.",
		})

		return
	}

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

	var user *dbcore.User
	userRequest := dbcore.DB.Where("email = ? OR username = ?", req.Email, req.Username).Find(&user)

	if userRequest.Error != nil {
		log.Warnf("failed to find if user exists or not: %s", userRequest.Error)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to find if user exists",
		})

		return
	}

	userExists := userRequest.RowsAffected > 0

	if userExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User already exists",
		})

		return
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		log.Warnf("Failed to generate password for client upon signup: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate password hash",
		})

		return
	}

	permissions := []dbcore.Permission{}

	for _, permission := range constants.DefaultPermissionNodes {
		permissionEnabledState := false

		if unsafeSignup || strings.HasPrefix(permission, "routes.") || permission == "permissions.see" {
			permissionEnabledState = true
		}

		permissions = append(permissions, dbcore.Permission{
			PermissionNode: permission,
			HasPermission:  permissionEnabledState,
		})
	}

	tokenRandomData := make([]byte, 80)

	if _, err := rand.Read(tokenRandomData); err != nil {
		log.Warnf("Failed to read random data to use as token: %s", err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})

		return
	}

	user = &dbcore.User{
		Email:       req.Email,
		Username:    req.Username,
		Name:        req.Name,
		IsBot:       &req.IsBot,
		Password:    base64.StdEncoding.EncodeToString(passwordHashed),
		Permissions: permissions,
		Tokens: []dbcore.Token{
			{
				Token:          base64.StdEncoding.EncodeToString(tokenRandomData),
				DisableExpiry:  forceNoExpiryTokens,
				CreationIPAddr: c.ClientIP(),
			},
		},
	}

	if result := dbcore.DB.Create(&user); result.Error != nil {
		log.Warnf("Failed to create user: %s", result.Error.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add user into database",
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
