package utilities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"github.com/hasura/go-graphql-client"
	"time"
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

func (s *libreDataStoreGraphQL) GetSubscription(q interface{}, vars map[string]interface{}) ports.LibreDataStoreSubscriptionPort {
	return newLibreDataStoreSubscriptionGraphQL(q, vars)
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

////////////////////////////////////////////////////////////////////////////////////////
type libreDataStoreSubscriptionGraphQL struct {
	libreLogger.LoggingEnabler
	libreConfig.ConfigurationEnabler
	queryStruct    interface{}
	queryVars      map[string]interface{}
	subClient      *graphql.SubscriptionClient
	subscriptionId string
	noticeChannel  chan []byte
}

func newLibreDataStoreSubscriptionGraphQL(q interface{}, vars map[string]interface{}) *libreDataStoreSubscriptionGraphQL {
	sub := libreDataStoreSubscriptionGraphQL{}
	sub.SetConfigCategory("libreDataStoreGraphQL")
	sub.SetLoggerConfigHook("graphQLSubs")
	sub.queryStruct = q
	sub.queryVars = vars
	return &sub
}

func (s *libreDataStoreSubscriptionGraphQL) SetSubscriptionQuery(q interface{}, vars map[string]interface{}) {
	s.queryStruct = q
	s.queryVars = vars
}

func (s *libreDataStoreSubscriptionGraphQL) GetSubscriptionNotifications(notificationChannel chan []byte) {
	s.LogDebugf("BEGIN GetSubscriptionNotifications")
	s.noticeChannel = notificationChannel
	url, _ := s.GetConfigItem("GRAPHQL_URL")

	s.subClient = graphql.NewSubscriptionClient(url).WithLog(s.LogDebug).
		WithoutLogTypes(graphql.GQL_DATA, graphql.GQL_CONNECTION_KEEP_ALIVE).
		OnError(func(sc *graphql.SubscriptionClient, err error) error {
			s.LogDebug("err", err)
			return err
		}).WithTimeout(time.Hour * 24)
	subid, err := s.subClient.Subscribe(s.queryStruct, s.queryVars, s.subscriptionHandler)
	if err == nil {
		s.LogDebugf("GetSubscriptionNotifications subscribed with id=%s", subid)
		s.subscriptionId = subid
		go s.subClient.Run()
	} else {
		panic(fmt.Sprintf("Error subscribing in graphQL client: %s", err))
	}
}

func (s *libreDataStoreSubscriptionGraphQL) StopGettingSubscriptionNotifications() {
	s.LogInfof("StopGettingSubscriptionNotifications begins for subscription id=%s", s.subscriptionId)
	err := s.subClient.Unsubscribe(s.subscriptionId)
	if err != nil {
		s.LogErrorf("Failed to unsubscribe from graphql subscription id=%s", s.subscriptionId)
	}
	err = s.subClient.Close()
	if err != nil {
		s.LogErrorf("Failed to close from graphql subscription id=%s", s.subscriptionId)
	}
}

func (s *libreDataStoreSubscriptionGraphQL) subscriptionHandler(message *json.RawMessage, err error) error {
	s.LogInfof("Handling a subscription notification in libreDataStoreSubscriptionGraphQL with err=%s msg=%s", err, *message)
	if err == nil {
		s.noticeChannel <- *message
	} else {
		s.LogErrorf("ERROR FROM SUBSCRIPTION QUERY: %s", err)
	}
	return nil
}
