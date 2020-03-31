package resolver

import (
	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/usecase/auth"
	"github.com/short-d/short/app/usecase/auth/payload"
	"github.com/short-d/short/app/usecase/requester"
	"github.com/short-d/short/app/usecase/url"
)

// Resolver contains GraphQL request handlers.
type Resolver struct {
	Query
	Mutation
}

type Factory struct {
}

// NewResolver creates a new GraphQL resolver.
func NewResolver(
	logger fw.Logger,
	tracer fw.Tracer,
	urlRetriever url.Retriever,
	urlCreator url.Creator,
	requesterVerifier requester.Verifier,
	authenticatorFactory auth.AuthenticatorFactory,
	payloadFactory payload.Factory,
) Resolver {
	authenticator := authenticatorFactory.MakeAuthenticator(payloadFactory)
	return Resolver{
		Query: newQuery(logger, tracer, authenticator, urlRetriever),
		Mutation: newMutation(
			logger,
			tracer,
			urlCreator,
			requesterVerifier,
			authenticator,
		),
	}
}
