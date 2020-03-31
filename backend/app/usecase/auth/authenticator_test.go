// +build !integration all

package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/short-d/short/app/usecase/auth/token"

	"github.com/short-d/app/fw"
	"github.com/short-d/app/mdtest"
	"github.com/short-d/short/app/entity"
	"github.com/short-d/short/app/usecase/auth/payload"
)

func TestAuthenticator_GenerateToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		issuedAt             string
		fromUserPayload      payload.Payload
		expectedTokenPayload fw.TokenPayload
	}{
		{
			name:     "empty payload",
			issuedAt: "2006-01-02T15:04:05Z",
			fromUserPayload: payload.Stub{
				TokenPayload: map[string]interface{}{},
			},
			expectedTokenPayload: map[string]interface{}{
				"issued_at": "2006-01-02T15:04:05Z",
			},
		},
		{
			name:     "payload contains ID and email",
			issuedAt: "2006-01-02T15:04:05Z",
			fromUserPayload: payload.Stub{
				TokenPayload: map[string]interface{}{
					"id":    "alpha",
					"email": "alpha@example.com",
				},
			},
			expectedTokenPayload: map[string]interface{}{
				"id":        "alpha",
				"email":     "alpha@example.com",
				"issued_at": "2006-01-02T15:04:05Z",
			},
		},
		{
			name:     "issue_at is override to current time",
			issuedAt: "2006-01-02T15:04:05Z",
			fromUserPayload: payload.Stub{
				TokenPayload: map[string]interface{}{
					"id":        "alpha",
					"email":     "alpha@example.com",
					"issued_at": "2001-02-05T15:04:05Z",
				},
			},
			expectedTokenPayload: map[string]interface{}{
				"id":        "alpha",
				"email":     "alpha@example.com",
				"issued_at": "2006-01-02T15:04:05Z",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tokenizer := mdtest.NewCryptoTokenizerFake()
			now, err := time.Parse(time.RFC3339, testCase.issuedAt)
			mdtest.Equal(t, nil, err)

			timer := mdtest.NewTimerFake(now)
			payloadFactory := payload.FactoryStub{
				Payload: testCase.fromUserPayload,
			}

			tokenIssuerFactory := token.NewIssuerFactory(tokenizer, timer)
			authenticatorFactory := NewAuthenticatorFactory(
				timer,
				2*time.Millisecond,
				tokenIssuerFactory,
			)
			authenticator := authenticatorFactory.MakeAuthenticator(payloadFactory)

			authToken, err := authenticator.GenerateToken(entity.User{})
			mdtest.Equal(t, nil, err)

			tokenPayload, err := tokenizer.Decode(authToken)
			mdtest.Equal(t, nil, err)
			mdtest.Equal(t, testCase.expectedTokenPayload, tokenPayload)
		})
	}
}

func TestAuthenticator_IsSignedIn(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name                  string
		expIssuedAt           time.Time
		tokenValidDuration    time.Duration
		currentTime           time.Time
		tokenPayload          fw.TokenPayload
		userInfoValidationErr error
		expIsSignIn           bool
	}{
		{
			name:               "empty token payload",
			expIssuedAt:        now,
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload:       map[string]interface{}{},
			expIsSignIn:        false,
		},
		{
			name:               "token expired",
			expIssuedAt:        now,
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(2 * time.Hour),
			tokenPayload: map[string]interface{}{
				"issued_at": now.Format(time.RFC3339),
			},
			expIsSignIn: false,
		},
		{
			name:               "token valid",
			expIssuedAt:        now,
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"issued_at": now.Format(time.RFC3339),
			},
			expIsSignIn: true,
		},
		{
			name:               "incorrect user info",
			expIssuedAt:        now,
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"email":     "alpha@example.com",
				"issued_at": now.Format(time.RFC3339),
			},
			userInfoValidationErr: errors.New("user ID not found in token payload"),
			expIsSignIn:           false,
		},
		{
			name:               "no issue_at in the payload",
			expIssuedAt:        now,
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"email": "alpha@example.com",
			},
			expIsSignIn: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tokenizer := mdtest.NewCryptoTokenizerFake()
			timer := mdtest.NewTimerFake(testCase.currentTime)
			payloadStub := payload.Stub{TokenPayload: testCase.tokenPayload}
			payloadFactory := payload.FactoryStub{
				Payload:  payloadStub,
				TokenErr: testCase.userInfoValidationErr,
			}
			tokenIssuerFactory := token.NewIssuerFactory(tokenizer, timer)
			authenticatorFactory := NewAuthenticatorFactory(
				timer,
				testCase.tokenValidDuration,
				tokenIssuerFactory,
			)
			authenticator := authenticatorFactory.MakeAuthenticator(payloadFactory)

			token, err := tokenizer.Encode(testCase.tokenPayload)
			mdtest.Equal(t, nil, err)
			gotIsSignIn := authenticator.IsSignedIn(token)
			mdtest.Equal(t, testCase.expIsSignIn, gotIsSignIn)
		})
	}
}

func TestAuthenticator_GetUser(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name                  string
		tokenValidDuration    time.Duration
		currentTime           time.Time
		tokenPayload          fw.TokenPayload
		user                  entity.User
		userInfoValidationErr error
		hasErr                bool
		expectedUser          entity.User
	}{
		{
			name:               "empty token payload",
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload:       map[string]interface{}{},
			hasErr:             true,
		},
		{
			name:               "token expired",
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(2 * time.Hour),
			tokenPayload: map[string]interface{}{
				"issued_at": now.Format(time.RFC3339),
			},
			hasErr: true,
		},
		{
			name:               "token valid",
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"issued_at": now.Format(time.RFC3339),
			},
			user: entity.User{
				ID:    "alpha",
				Name:  "Alpha",
				Email: "alpha@example.com",
			},
			hasErr: false,
			expectedUser: entity.User{
				ID:    "alpha",
				Name:  "Alpha",
				Email: "alpha@example.com",
			},
		},
		{
			name:               "incorrect user info",
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"email":     "alpha@example.com",
				"issued_at": now.Format(time.RFC3339),
			},
			userInfoValidationErr: errors.New("user ID not found in token payload"),
			hasErr:                true,
		},
		{
			name:               "no issue_at in the payload",
			tokenValidDuration: time.Hour,
			currentTime:        now.Add(30 * time.Minute),
			tokenPayload: map[string]interface{}{
				"email": "alpha@example.com",
			},
			hasErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tokenizer := mdtest.NewCryptoTokenizerFake()
			timer := mdtest.NewTimerFake(testCase.currentTime)
			tokenPayloadStub := payload.Stub{
				TokenPayload: testCase.tokenPayload,
				User:         testCase.user,
			}
			payloadFactory := payload.FactoryStub{
				Payload:  tokenPayloadStub,
				TokenErr: testCase.userInfoValidationErr,
			}
			tokenIssuerFactory := token.NewIssuerFactory(tokenizer, timer)
			authenticatorFactory := NewAuthenticatorFactory(
				timer,
				testCase.tokenValidDuration,
				tokenIssuerFactory,
			)
			authenticator := authenticatorFactory.MakeAuthenticator(payloadFactory)

			token, err := tokenizer.Encode(testCase.tokenPayload)
			mdtest.Equal(t, nil, err)
			gotUser, err := authenticator.GetUser(token)
			if testCase.hasErr {
				mdtest.NotEqual(t, nil, err)
				return
			}
			mdtest.Equal(t, testCase.expectedUser, gotUser)
		})
	}
}
