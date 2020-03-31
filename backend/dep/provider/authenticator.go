package provider

import (
	"time"

	"github.com/short-d/short/app/usecase/auth/token"

	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/usecase/auth"
)

// TokenValidDuration represents the duration of a valid token.
type TokenValidDuration time.Duration

// NewAuthenticator creates Authenticator with TokenValidDuration to uniquely identify duration during dependency injection.
func NewAuthenticatorFactory(
	timer fw.Timer,
	duration TokenValidDuration,
	tokenIssuerFactory token.IssuerFactory,
) auth.AuthenticatorFactory {
	return auth.NewAuthenticatorFactory(timer, time.Duration(duration), tokenIssuerFactory)
}
