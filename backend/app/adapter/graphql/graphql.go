package graphql

import (
	"github.com/short-d/app/fw"
	"github.com/short-d/short/app/adapter/graphql/resolver"
)

var _ fw.GraphQLAPI = (*Short)(nil)

// Short represents GraphQL API config
type Short struct {
	resolver *resolver.Resolver
}

// GetSchema retrieves GraphQL schema
func (t Short) GetSchema() string {
	return schema
}

// GetResolver retrieves GraphQL resolver
func (t Short) GetResolver() fw.Resolver {
	return t.resolver
}

// NewShort creates GraphQL API config
func NewShort(
	//logger fw.Logger,
	//tracer fw.Tracer,
	//urlRetriever url.Retriever,
	//urlCreator url.Creator,
	//requesterVerifier requester.Verifier,
	graphQLResolver resolver.Resolver,
	//authenticator auth.Authenticator,
) Short {
	//r := resolver.NewResolver(
	//	logger,
	//	tracer,
	//	urlRetriever,
	//	urlCreator,
	//	requesterVerifier,
	//	authenticator,
	//)
	return Short{
		resolver: &graphQLResolver,
	}
}
