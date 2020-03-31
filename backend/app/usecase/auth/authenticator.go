package auth

import (
	"errors"
	"time"

	"github.com/short-d/short/app/usecase/auth/token"

	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/entity"
	"github.com/short-d/short/app/usecase/auth/payload"
)

// Authenticator securely authenticates an user's identity.
type Authenticator struct {
	timer              fw.Timer
	tokenValidDuration time.Duration
	tokenIssuer        token.Issuer
}

// IsSignedIn checks whether user successfully signed in
func (a Authenticator) IsSignedIn(tokenString string) bool {
	authToken, err := a.tokenIssuer.ParseToken(tokenString)
	if err != nil {
		return false
	}
	return a.isTokenValid(authToken)
}

// GetUser decodes authentication token to user data
func (a Authenticator) GetUser(tokenString string) (entity.User, error) {
	authToken, err := a.tokenIssuer.ParseToken(tokenString)
	if err != nil {
		return entity.User{}, err
	}

	if !a.isTokenValid(authToken) {
		return entity.User{}, errors.New("token expired")
	}
	return authToken.User, nil
}

// GenerateToken encodes part of user data into authentication token
func (a Authenticator) GenerateToken(user entity.User) (string, error) {
	return a.tokenIssuer.IssuedToken(user)
}

func (a Authenticator) isTokenValid(authToken token.Token) bool {
	tokenExpireAt := authToken.IssuedAt.Add(a.tokenValidDuration)
	now := a.timer.Now()
	return !tokenExpireAt.Before(now)
}

type AuthenticatorFactory struct {
	tokenIssuerFactory token.IssuerFactory
	timer              fw.Timer
	tokenValidDuration time.Duration
}

func (a AuthenticatorFactory) MakeAuthenticator(
	payloadFactory payload.Factory,
) Authenticator {
	tokenIssuer := a.tokenIssuerFactory.MakeIssuer(payloadFactory)
	return Authenticator{
		tokenIssuer:        tokenIssuer,
		timer:              a.timer,
		tokenValidDuration: a.tokenValidDuration,
	}
}

// NewAuthenticator initializes authenticator with custom token valid duration
func NewAuthenticatorFactory(
	timer fw.Timer,
	tokenValidDuration time.Duration,
	tokenIssuerFactory token.IssuerFactory,
) AuthenticatorFactory {
	return AuthenticatorFactory{
		timer:              timer,
		tokenValidDuration: tokenValidDuration,
		tokenIssuerFactory: tokenIssuerFactory,
	}
}
