package apiclient

import "git.terah.dev/imterah/hermes/apiclient/users"

type HermesAPIClient struct {
	URL string
}

/// Users

func (api *HermesAPIClient) UserGetRefreshToken(username *string, email *string, password string) (string, error) {
	return users.GetRefreshToken(api.URL, username, email, password)
}

func (api *HermesAPIClient) UserGetJWTFromToken(refreshToken string) (string, error) {
	return users.GetJWTFromToken(api.URL, refreshToken)
}

func (api *HermesAPIClient) UserCreate(fullName, username, email, password string, isBot bool) (string, error) {
	return users.CreateUser(api.URL, fullName, username, email, password, isBot)
}
