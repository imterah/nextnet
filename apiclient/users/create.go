package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.terah.dev/imterah/hermes/apiclient/backendstructs"
)

type createUserResponse struct {
	Error        string `json:"error"`
	Success      bool   `json:"success"`
	RefreshToken string `json:"refreshToken"`
}

func CreateUser(url, fullName, username, email, password string, isBot bool) (string, error) {
	body, err := json.Marshal(&backendstructs.UserCreationRequest{
		Username: username,
		Name:     fullName,
		Email:    email,
		Password: password,
		IsBot:    isBot,
	})

	if err != nil {
		return "", err
	}

	res, err := http.Post(fmt.Sprintf("%s/api/v1/users/create", url), "application/json", bytes.NewBuffer(body))

	if err != nil {
		return "", err
	}

	bodyContents, err := io.ReadAll(res.Body)

	if err != nil {
		return "", fmt.Errorf("failed to read response body: %s", err.Error())
	}

	response := &createUserResponse{}

	if err := json.Unmarshal(bodyContents, response); err != nil {
		return "", err
	}

	if response.Error != "" {
		return "", fmt.Errorf("error from server: %s", response.Error)
	}

	if !response.Success {
		return "", fmt.Errorf("failed to get refresh token")
	}

	if response.RefreshToken == "" {
		return "", fmt.Errorf("refresh token is empty")
	}

	return response.RefreshToken, nil
}
