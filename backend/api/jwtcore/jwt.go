package jwtcore

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTKey []byte

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

	return nil
}

func Parse(tokenString string, options ...jwt.ParserOption) (*jwt.Token, error) {
	return jwt.Parse(tokenString, JWTKeyCallback, options...)
}

func Generate(uid uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Minute)),
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
