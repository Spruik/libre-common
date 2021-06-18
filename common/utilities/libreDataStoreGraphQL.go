package utilities

import (
	"context"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"github.com/hasura/go-graphql-client"
)

type libreDataStoreGraphQL struct {
	//inherit config functions
	libreConfig.ConfigurationEnabler

	//inherit logging functions
	libreLogger.LoggingEnabler

	gqlClient *graphql.Client
}

func NewLibreDataStoreGraphQL() *libreDataStoreGraphQL {

	s := libreDataStoreGraphQL{}
	s.SetConfigCategory("libreDataStoreGraphQL")
	s.SetLoggerConfigHook("DATAGQL")
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating a graphQL client
func (s *libreDataStoreGraphQL) Connect() error {
	url, _ := s.GetConfigItem("GRAPHQL_URL")
	s.gqlClient = graphql.NewClient(url, nil)

	s.LogInfof("GraphQL Data Store client created for %s", url)
	return nil
}

//Close implements the interface by closing the graphQL client
func (s *libreDataStoreGraphQL) Close() error {
	s.LogInfof("GraphQL Data Store client closed")
	return nil
}

func (s *libreDataStoreGraphQL) BeginTransaction(forUpdate bool, name string) ports.LibreDataStoreTransactionPort {
	return newLibreDataStoreTransactionGraphQL(s, forUpdate, name)
}

////////////////////////////////////////////////////////////////////////////////////////
type libreDataStoreTransactionGraphQL struct {
	libreLogger.LoggingEnabler
	txnName string
	//no transactions, so we just use the client
	gqlClient *graphql.Client
	forUpdate bool
}

func newLibreDataStoreTransactionGraphQL(s *libreDataStoreGraphQL, forUpdate bool, name string) *libreDataStoreTransactionGraphQL {

	t := libreDataStoreTransactionGraphQL{
		gqlClient: s.gqlClient,
		txnName:   name,
		forUpdate: forUpdate,
	}
	t.SetLoggerConfigHook("graphQLTrans")
	return &t
}

func (s *libreDataStoreTransactionGraphQL) ExecuteQuery(q interface{}, vars map[string]interface{}) error {
	s.LogDebugf("GraphQL Transaction %s doing Query %+v (%+v)", s.txnName, q, vars)
	err := s.gqlClient.Query(context.Background(), q, vars)
	s.LogDebugf("GraphQL Transaction %s done with Query", s.txnName)
	return err
}

func (s *libreDataStoreTransactionGraphQL) ExecuteMutation(m interface{}, vars map[string]interface{}) error {
	s.LogDebugf("GraphQL Transaction %s doing Mutate", s.txnName)
	err := s.gqlClient.Mutate(context.Background(), m, vars)
	s.LogDebugf("GraphQL Transaction %s done with Mutate", s.txnName)
	return err
}

func (s *libreDataStoreTransactionGraphQL) Commit() {
	//noop
}
func (s *libreDataStoreTransactionGraphQL) Dispose() {
	//noop
}
