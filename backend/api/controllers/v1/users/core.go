package users

import "os"

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
