package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.terah.dev/imterah/hermes/apiclient/backendstructs"
)

type refreshTokenResponse struct {
	Success      bool   `json:"success"`
	RefreshToken string `json:"refreshToken"`
}

type jwtTokenResponse struct {
	Success bool   `json:"success"`
	JWT     string `json:"token"`
}

func GetRefreshToken(url string, username, email *string, password string) (string, error) {
	body, err := json.Marshal(&backendstructs.UserLoginRequest{
		Username: username,
		Email:    email,
		Password: password,
	})

	if err != nil {
		return "", err
	}

	res, err := http.Post(fmt.Sprintf("%s/api/v1/users/login", url), "application/json", bytes.NewBuffer(body))

	if err != nil {
		return "", err
	}

	bodyContents, err := io.ReadAll(res.Body)

	if err != nil {
		return "", fmt.Errorf("failed to read response body: %s", err.Error())
	}

	response := &refreshTokenResponse{}

	if err := json.Unmarshal(bodyContents, response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get refresh token")
	}

	if response.RefreshToken == "" {
		return "", fmt.Errorf("refresh token is empty")
	}

	return response.RefreshToken, nil
}

func GetJWTFromToken(url, refreshToken string) (string, error) {
	body, err := json.Marshal(&backendstructs.UserRefreshRequest{
		Token: refreshToken,
	})

	if err != nil {
		return "", err
	}

	res, err := http.Post(fmt.Sprintf("%s/api/v1/users/refresh", url), "application/json", bytes.NewBuffer(body))

	if err != nil {
		return "", err
	}

	bodyContents, err := io.ReadAll(res.Body)

	if err != nil {
		return "", fmt.Errorf("failed to read response body: %s", err.Error())
	}

	response := &jwtTokenResponse{}

	if err := json.Unmarshal(bodyContents, response); err != nil {
		return "", err
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get JWT token")
	}

	if response.JWT == "" {
		return "", fmt.Errorf("JWT token is empty")
	}

	return response.JWT, nil
}
