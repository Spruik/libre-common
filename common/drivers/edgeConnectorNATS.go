package drivers

import (
	"fmt"
	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/nats-io/nats.go"
)

type edgeConnectorNATS struct {
	//inherit logging functions
	libreLogger.LoggingEnabler
	//inherit config functions
	libreConfig.ConfigurationEnabler

	natsConn       *nats.Conn
	ChangeChannels map[string]chan domain.StdMessageStruct
	singleChannel  chan domain.StdMessageStruct

	topicTemplate   string
	tagDataCategory string
	eventCategory   string

	// Keep track of client topics so we can stop listening/unsubscribe by client
	clientTopics map[string][]*nats.Subscription
}

//Initialize edge connectors for NATS
func NewEdgeConnectorNATS() *edgeConnectorNATS {
	s := edgeConnectorNATS{natsConn: nil}
	s.SetConfigCategory("edgeConnectorNATS")
	s.SetLoggerConfigHook("EDGENATS")
	s.ChangeChannels = make(map[string]chan domain.StdMessageStruct)
	s.topicTemplate, _ = s.GetConfigItemWithDefault("TOPIC_TEMPLATE", "<EQNAME>/Report/<TAGNAME>")
	s.tagDataCategory, _ = s.GetConfigItemWithDefault("TAG_DATA_CATEGORY", "EdgeTagChange")
	s.eventCategory, _ = s.GetConfigItemWithDefault("EVENT_CATEGORY", "EdgeEvent")
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an NATS client

func (s *edgeConnectorNATS) Connect(connInfo map[string]interface{}) error {
	var err error
	var server string
	if server, err = s.GetConfigItem("NATS_SERVER"); err == nil {
		s.LogInfo("Found all the config items")
	}
	if err != nil {
		panic("Failed to find configuration data for libreConnectorNATS")
	}

	//Establish NATS Connection
	natsConn, err := nats.Connect(server)
	s.natsConn = natsConn
	if err != nil {
		panic("Failed to connect to NATS server.")
	}
	s.LogInfof("NATS Connected to %s", server)
	return err
}

//Close implements the interface by closing the NATS client
func (s *edgeConnectorNATS) Close() error {
	if s.natsConn == nil {
		return nil
	}
	s.natsConn.Close()
	s.LogInfof("Edge Connection Closed\n")
	return nil
}

//ReadTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorNATS) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorNATS) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

func (s *edgeConnectorNATS) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
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
		topic := s.buildTopicString(val)
		err := s.SubscribeToTopic(clientName, topic)
		if err == nil {
			s.LogInfof("Subscribed to topic %s", topic)
		} else {
			panic(err)
		}
	}
}
func (s *edgeConnectorNATS) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - how to get history via MQTT - seems like it will depend on what is publishing the MQTT messages
	return []domain.StdMessageStruct{}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *edgeConnectorNATS) SubscribeToTopic(client string, topic string) error {
	js, err := s.natsConn.JetStream()
	if err != nil {
		panic("Failed to obtain JetStream")
	}
	subscription, err := js.Subscribe(topic, func(msg *nats.Msg) {
		// put the message data into the ItemValue field of the StdMessageStruct
		// ToDo: Can we use a template to decode the Subject string?
		// expected subjects are Libre.subscriptions.workflowSpecification and
		// Libre.events.task or Libre.commands.workflowSpecification
		// put the third level of the subject into the category field.
		subjects := strings.Split(msg.Subject, ".")
		s.LogInfo(msg.Data)
		var tagStruct domain.StdMessageStruct
		if len(subjects) < 3 {
			s.LogInfo("Category not found in subject string " + msg.Subject)
		} else {
			tagStruct.Category = subjects[1]    // events, commands or subscriptions
			tagStruct.OwningAsset = subjects[2] // workflowSpecification or task
		}
		// The fourth level in the subject is the event/command type. map this into the ItemDataType
		if len(subjects) >= 4 {
			s.LogInfo("Category not found in subject string " + msg.Subject)
			tagStruct.ItemDataType = subjects[3] // workflowSpecification or task
		}
		// The fifth level in the subject is the Id. map this into the OwningAssetId
		if len(subjects) >= 4 {
			s.LogInfo("Category not found in subject string " + msg.Subject)
			tagStruct.OwningAssetId = subjects[4] // workflowSpecification or task
		}
		tagStruct.ItemValue = string(msg.Data)
		if err == nil {
			if s.singleChannel == nil {
				s.ChangeChannels[tagStruct.OwningAsset] <- tagStruct
			} else {
				s.singleChannel <- tagStruct
			}
		} else {
			s.LogErrorf("Failed to unmarshal the payload of the incoming message: %s [%s]", string(msg.Data), err)
		}
		msg.Ack()
	})

	// Store so we can later unsubscribe
	s.clientTopics[client] = append(s.clientTopics[client], subscription)

	return err
}

func (s *edgeConnectorNATS) buildTopicString(changeFilterVal interface{}) string {
	topic := fmt.Sprintf("%s", changeFilterVal)
	return topic
}

// StopListeningForTagChanges for a given topic
func (s *edgeConnectorNATS) StopListeningForTagChanges(client string) (err error) {
	subscriptions, exists := s.clientTopics[client]

	// Return early if no subscriptions for the client
	if !exists {
		return nil
	}

	// Unsubscribe
	for _, subscription := range subscriptions {
		err = subscription.Unsubscribe()
		if err != nil {
			s.LogDebugf("edgeConnectorNATS unsubscribed from %s expected no error; got %s", subscription.Subject, err)
		}
	}

	// Remove map entry in subscribtions
	delete(s.clientTopics, client)

	// Remove map entry in message channel
	delete(s.ChangeChannels, client)

	return err
}
