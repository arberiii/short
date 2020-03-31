// +build !integration all

package resolver

import (
	"testing"
	"time"

	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/usecase/auth/payload"
	"github.com/short-d/short/app/usecase/auth/token"

	"github.com/short-d/app/mdtest"
	"github.com/short-d/short/app/entity"
	"github.com/short-d/short/app/usecase/auth"
	"github.com/short-d/short/app/usecase/repository"
	"github.com/short-d/short/app/usecase/url"
)

func TestQuery_AuthQuery(t *testing.T) {
	user := entity.User{
		Email: "alpha@example.com",
	}
	payloadStub := payload.Stub{
		TokenPayload: fw.TokenPayload{
			"email": "alpha@example.com",
		},
		User: user,
	}
	payloadFactory := payload.FactoryStub{
		Payload: payloadStub,
	}
	authenticator := newAuthenticator(time.Now(), time.Hour, payloadFactory)

	authToken, err := authenticator.GenerateToken(user)
	mdtest.Equal(t, nil, err)
	randomToken := "random_token"

	testCases := []struct {
		name      string
		authToken *string
		expHasErr bool
		expUser   *entity.User
	}{
		{
			name:      "with valid auth token",
			authToken: &authToken,
			expHasErr: false,
			expUser: &entity.User{
				Email: "alpha@example.com",
			},
		},
		{
			name:      "with invalid auth token",
			authToken: &randomToken,
			expHasErr: true,
		},
		{
			name:      "without auth token",
			authToken: nil,
			expHasErr: false,
			expUser:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sqlDB, _, err := mdtest.NewSQLStub()
			mdtest.Equal(t, nil, err)
			defer sqlDB.Close()

			fakeRepo := repository.NewURLFake(map[string]entity.URL{})
			retrieverFake := url.NewRetrieverPersist(&fakeRepo)
			logger := mdtest.NewLoggerFake(mdtest.FakeLoggerArgs{})
			tracer := mdtest.NewTracerFake()
			query := newQuery(&logger, &tracer, authenticator, retrieverFake)

			mdtest.Equal(t, nil, err)
			authQueryArgs := AuthQueryArgs{AuthToken: testCase.authToken}
			authQuery, err := query.AuthQuery(&authQueryArgs)
			if testCase.expHasErr {
				mdtest.NotEqual(t, nil, err)
				return
			}
			mdtest.Equal(t, nil, err)
			mdtest.Equal(t, testCase.expUser, authQuery.user)
		})
	}
}

func newAuthenticator(
	now time.Time,
	tokenValidDuration time.Duration,
	payloadFactory payload.Factory,
) auth.Authenticator {
	tokenizer := mdtest.NewCryptoTokenizerFake()
	timer := mdtest.NewTimerFake(now)
	tokenIssuerFactory := token.NewIssuerFactory(tokenizer, timer)
	authFactory := auth.NewAuthenticatorFactory(
		timer,
		tokenValidDuration,
		tokenIssuerFactory,
	)
	return authFactory.MakeAuthenticator(payloadFactory)
}
