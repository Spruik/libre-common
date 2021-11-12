package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/drivers/gql"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"

	//"github.com/nats-io/nats.go"
	"log"

	//"strings"
	"time"
)

type edgeConnectorGraphQL struct {
	//inherit logging functions
	libreLogger.LoggingEnabler
	//inherit config functions
	libreConfig.ConfigurationEnabler

	gqlClient      *gql.SubscriptionClient
	ChangeChannels map[string]chan domain.StdMessageStruct
	singleChannel  chan domain.StdMessageStruct
	subscriptions  map[string]*domain.DataSubscription

	// Keep track of client topics so we can stop listening/unsubscribe by client
	clientTopics map[string][]string
}

//Initialize edge connectors for NATS
func NewEdgeConnectorGraphQL() *edgeConnectorGraphQL {
	s := edgeConnectorGraphQL{}
	s.SetConfigCategory("edgeConnectorGraphQL")
	s.SetLoggerConfigHook("EDGENATS")
	s.ChangeChannels = make(map[string]chan domain.StdMessageStruct)
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an NATS client

func (s *edgeConnectorGraphQL) Connect(connInfo map[string]interface{}) error {
	url, _ := s.GetConfigItem("GRAPHQL_URL")
	s.subscriptions = make(map[string]*domain.DataSubscription)
	s.gqlClient = gql.NewSubscriptionClient(url).
		WithConnectionParams(map[string]interface{}{
			"headers": map[string]string{
				"X-Auth0-Token": "",
			},
		}).WithLog(log.Println).
		WithTimeout(8760*time.Hour).
		WithoutLogTypes(gql.GQL_DATA, gql.GQL_CONNECTION_KEEP_ALIVE).
		OnError(func(sc *gql.SubscriptionClient, err error) error {
			log.Print("[ERROR]", err)
			return err
		})
	//defer s.gqlClient.Close()
	go func() {
		err := s.gqlClient.Run()
		s.LogError(err)
	}()
	s.LogInfof("GraphQL Data Store client created for %s", url)
	return nil
}

//Close implements the interface by closing the NATS client
func (s *edgeConnectorGraphQL) Close() error {
	s.LogInfof("GraphQL Subscription Client client closed")
	return nil
}

//ReadTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorGraphQL) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorGraphQL) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

func (s *edgeConnectorGraphQL) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	s.LogDebugf("BEGIN ListenForEdgeTagChanges")
	var clientName string
	fltrClient, exists := changeFilter["Client"]
	if exists {
		clientName = fmt.Sprintf("%s", fltrClient)
	} else {
		clientName = ""
	}
	if clientName == "" {
		if s.singleChannel == nil {
			s.singleChannel = c
		} else {
			panic("Cannot use more than one single channel listen")
		}
	} else {
		if s.singleChannel == nil {
			s.ChangeChannels[clientName] = c
		} else {
			panic("Cannot single channel listen with client-based listen")
		}
	}
	for _, val := range changeFilter {
		sub := s.buildSubscription(val)
		err := s.subscribe(clientName, sub)
		if err == nil {
			s.LogInfof("Subscribed to topic %s", sub.Topic)
		} else {
			panic(err)
		}

	}
}
func (s *edgeConnectorGraphQL) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - how to get history via MQTT - seems like it will depend on what is publishing the MQTT messages
	return []domain.StdMessageStruct{}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *edgeConnectorGraphQL) subscribe(client string, sub *domain.DataSubscription) error {
	var variables map[string]interface{}
	id, err := s.gqlClient.Subscribe(sub.Query, variables, func(data *json.RawMessage, err error) error {
		if err != nil {
			return err
		}
		j, _ := json.Marshal(data)
		s.LogDebug("Received msg : " + string(j))
		msg := domain.StdMessageStruct{
			OwningAsset:   sub.Topic,
			OwningAssetId: "",
			ItemName:      "",
			ItemNameExt:   nil,
			ItemId:        "",
			ItemValue:     string(j),
			//ItemOldValue:  "",
			ItemDataType: "",
			TagQuality:   0,
			Err:          nil,
			//ChangedTime:   time.Time{},
			Category: "EVENT",
		}
		if s.singleChannel == nil {
			s.ChangeChannels[msg.OwningAsset] <- msg
		} else {
			s.singleChannel <- msg
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	} else {
		sub.Id = id
		s.subscriptions[sub.Topic] = sub
		s.clientTopics[client] = append(s.clientTopics[client], id)
	}
	return nil
}

func (s *edgeConnectorGraphQL) buildSubscription(changeFilterVal interface{}) *domain.DataSubscription {
	sub := changeFilterVal.(*domain.DataSubscription)
	return sub
}

func (s *edgeConnectorGraphQL) StopListeningForTagChanges(client string) (err error) {
	subscriptionIDs, exists := s.clientTopics[client]

	// Return early if client doesn't exist
	if !exists {
		return nil
	}

	// Unsubscribe
	for _, subscriptionID := range subscriptionIDs {
		err = s.gqlClient.Unsubscribe(subscriptionID)
		if err != nil {
			s.LogDebugf("edgeConnectorGraphQL unsubscribed to %s expected no error; got %s", subscriptionID, err)
		}
		delete(s.subscriptions, subscriptionID)
	}

	// Remove from Client Map
	delete(s.clientTopics, client)

	// Remove from data channel
	delete(s.ChangeChannels, client)

	return err
}
