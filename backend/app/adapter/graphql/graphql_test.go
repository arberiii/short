// +build !integration all

package graphql

import (
	"testing"
	"time"

	"github.com/short-d/short/app/adapter/graphql/resolver"
	"github.com/short-d/short/app/usecase/auth/payload"
	"github.com/short-d/short/app/usecase/auth/token"

	"github.com/short-d/app/mdtest"
	"github.com/short-d/short/app/adapter/db"
	"github.com/short-d/short/app/usecase/auth"
	"github.com/short-d/short/app/usecase/keygen"
	"github.com/short-d/short/app/usecase/requester"
	"github.com/short-d/short/app/usecase/service"
	"github.com/short-d/short/app/usecase/url"
	"github.com/short-d/short/app/usecase/validator"
)

func TestGraphQlAPI(t *testing.T) {
	sqlDB, _, err := mdtest.NewSQLStub()
	mdtest.Equal(t, nil, err)
	defer sqlDB.Close()

	urlRepo := db.NewURLSql(sqlDB)
	retriever := url.NewRetrieverPersist(urlRepo)
	urlRelationRepo := db.NewUserURLRelationSQL(sqlDB)
	keyFetcher := service.NewKeyFetcherFake([]service.Key{})
	keyGen, err := keygen.NewKeyGenerator(2, &keyFetcher)
	mdtest.Equal(t, nil, err)

	longLinkValidator := validator.NewLongLink()
	customAliasValidator := validator.NewCustomAlias()
	creator := url.NewCreatorPersist(
		urlRepo,
		urlRelationRepo,
		keyGen,
		longLinkValidator,
		customAliasValidator,
	)

	s := service.NewReCaptchaFake(service.VerifyResponse{})
	verifier := requester.NewVerifier(s)
	authFactory := newAuthenticatorFactory(time.Now(), time.Hour)

	logger := mdtest.NewLoggerFake(mdtest.FakeLoggerArgs{})
	tracer := mdtest.NewTracerFake()
	payloadFactory := payload.FactoryStub{Payload: payload.Stub{}}
	graphQLResolver := resolver.NewResolver(
		&logger,
		&tracer,
		retriever,
		creator,
		verifier,
		authFactory,
		payloadFactory,
	)
	graphqlAPI := NewShort(graphQLResolver)
	mdtest.Equal(t, true, mdtest.IsGraphQlAPIValid(graphqlAPI))
}

func newAuthenticatorFactory(now time.Time, tokenValidDuration time.Duration) auth.AuthenticatorFactory {
	tokenizer := mdtest.NewCryptoTokenizerFake()
	timer := mdtest.NewTimerFake(now)
	tokenIssuerFactory := token.NewIssuerFactory(tokenizer, timer)
	return auth.NewAuthenticatorFactory(timer, tokenValidDuration, tokenIssuerFactory)
}
