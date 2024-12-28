package jwtcore

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"git.terah.dev/imterah/hermes/backend/api/dbcore"
	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTKey          []byte
	developmentMode bool
)

func SetupJWT() error {
	var err error
	jwtDataString := os.Getenv("HERMES_JWT_SECRET")

	if jwtDataString == "" {
		return fmt.Errorf("JWT secret isn't set (missing HERMES_JWT_SECRET)")
	}

	if os.Getenv("HERMES_JWT_BASE64_ENCODED") != "" {
		JWTKey, err = base64.StdEncoding.DecodeString(jwtDataString)

		if err != nil {
			return fmt.Errorf("failed to decode base64 JWT: %s", err.Error())
		}
	} else {
		JWTKey = []byte(jwtDataString)
	}

	if os.Getenv("HERMES_DEVELOPMENT_MODE") != "" {
		developmentMode = true
	}

	return nil
}

func Parse(tokenString string, options ...jwt.ParserOption) (*jwt.Token, error) {
	return jwt.Parse(tokenString, JWTKeyCallback, options...)
}

func GetUserFromJWT(token string) (*dbcore.User, error) {
	parsedJWT, err := Parse(token)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("token is expired")
		} else {
			return nil, err
		}
	}

	audience, err := parsedJWT.Claims.GetAudience()

	if err != nil {
		return nil, err
	}

	if len(audience) < 1 {
		return nil, fmt.Errorf("audience is too small")
	}

	uid, err := strconv.Atoi(audience[0])

	if err != nil {
		return nil, err
	}

	user := &dbcore.User{}
	userRequest := dbcore.DB.Preload("Permissions").Where("id = ?", uint(uid)).Find(&user)

	if userRequest.Error != nil {
		return user, fmt.Errorf("failed to find if user exists or not: %s", userRequest.Error.Error())
	}

	userExists := userRequest.RowsAffected > 0

	if !userExists {
		return user, fmt.Errorf("user does not exist")
	}

	return user, nil
}

func Generate(uid uint) (string, error) {
	timeMultiplier := 3

	if developmentMode {
		timeMultiplier = 60 * 24
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(timeMultiplier) * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Audience:  []string{strconv.Itoa(int(uid))},
	})

	signedToken, err := token.SignedString(JWTKey)

	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func JWTKeyCallback(*jwt.Token) (interface{}, error) {
	return JWTKey, nil
}
