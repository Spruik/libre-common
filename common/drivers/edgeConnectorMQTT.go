package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/drivers/autopaho"

	"log"

	"net/url"
	"os"
	"regexp"

	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/eclipse/paho.golang/paho"
)

const ERROR_MESSAGE_FAILED_TO_CONNECT = "Failed to connect to %s: %s"

type edgeConnectorMQTT struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	mqttConnectionManager     *autopaho.ConnectionManager
	mqttClient     *paho.Client

	ChangeChannels map[string]chan domain.StdMessageStruct
	singleChannel  chan domain.StdMessageStruct
	config         map[string]string

	topicTemplate    string
	topicParseRegExp *regexp.Regexp
	tagDataCategory  string
	eventCategory    string
}

func NewEdgeConnectorMQTT(configHook string) *edgeConnectorMQTT {
	s := edgeConnectorMQTT{
		mqttClient: nil,
	}
	s.ChangeChannels = make(map[string]chan domain.StdMessageStruct)
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	s.topicTemplate, _ = s.GetConfigItemWithDefault("TOPIC_TEMPLATE", "<EQNAME>/Report/<TAGNAME>")
	s.tagDataCategory, _ = s.GetConfigItemWithDefault("TAG_DATA_CATEGORY", "EdgeTagChange")
	s.eventCategory, _ = s.GetConfigItemWithDefault("EVENT_CATEGORY", "EdgeEvent")
	topicRE := s.topicTemplate
	topicRE = strings.Replace(topicRE, "<EQNAME>", "(?P<EQNAME>[A-Za-z0-9_\\/]*)", -1)
	topicRE = strings.Replace(topicRE, "<TAGNAME>", "(?P<TAGNAME>[A-Za-z0-9_]*)", -1)
	s.topicParseRegExp = regexp.MustCompile(topicRE)
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *edgeConnectorMQTT) Connect(clientId string) error {
	var err error
	var server, user, pwd , svcName string

	//Grab server address config
	server, err = s.GetConfigItem("MQTT_SERVER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SERVER: " + server)
	} else {
		s.LogError("Config read failed:  MQTT_SERVER" , err)
		panic("edgeConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab password
	pwd, err = s.GetConfigItem("MQTT_PWD")
	if err == nil {
		s.LogDebug("Config found:  MQTT_PWD: <will not be shown in log>")
	} else {
		s.LogError("Config read failed:  MQTT_PWD" , err)
		panic("edgeConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab user
	user, err = s.GetConfigItem("MQTT_USER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_USER: " + user)
	} else {
		s.LogError("Config read failed:  MQTT_USER" , err)
		panic("edgeConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab service name
	svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SVC_NAME: " + svcName)
	} else {
		s.LogError("Config read failed:  MQTT_SVC_NAME" , err)
		panic("edgeConnectorMQTT failed to find configuration data for MQTT connection")
	}

	MQTTServerURL, err := url.Parse(server)

	if err == nil {
		s.LogDebugf("URL Schema (eg:  mqtt:// ) determines the connection type used by Libre")
		lowerCaseScheme := strings.ToLower(MQTTServerURL.Scheme)
		switch  lowerCaseScheme{
		case "mqtt", "tcp":
			s.LogDebugf("URL Scheme: [%s] requires INSECURE connection, Libre edge will not connect if the MQTT broker has TLS enabled" , lowerCaseScheme)
		case "ssl", "tls", "mqtts", "mqtt+ssl", "tcps":
			s.LogDebugf("URL Scheme: [%s] requires SECURE connection, TLS is required at the MQTT broker" , lowerCaseScheme)
		default:
			s.LogErrorf("URL Scheme: [%s] is not supported" , lowerCaseScheme)
			s.LogErrorf("Supported SECURE schema are: ssl, tls, mqtts, mqtt+ssl, tcps")
			s.LogErrorf("Other supported schema are: mqtt, tcp")
			panic("edgeConnectorMQTT server name specifies unsupported URL scheme")
		}
	} else {
		s.LogError("Server name not valid" , err)
		panic("edgeConnectorMQTT server name not valid")
	}

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{MQTTServerURL},
		KeepAlive:         300,
		ConnectRetryDelay: 10 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			s.LogInfo("mqtt connection up")
		},
		OnConnectError: func(err error) { s.LogError("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: clientId,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				s.tagChangeHandler(m)
			}),
			OnClientError: func(err error) { s.LogError("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					s.LogWarnf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					s.LogWarnf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}
	cliCfg.Debug = log.New(os.Stdout,"autoPaho",1)
	cliCfg.PahoDebug = log.New(os.Stdout,"paho",1)
	cliCfg.SetUsernamePassword(user,[]byte(pwd))
	ctx, _ := context.WithCancel(context.Background())
	cm, err := autopaho.NewConnection(ctx, cliCfg)
	err = cm.AwaitConnection(ctx)
	s.mqttConnectionManager=cm
	return err
}

//Close implements the interface by closing the MQTT client
func (s *edgeConnectorMQTT) Close() error {
	//if s.mqttClient == nil {
	//	return nil
	//}
	//disconnStruct := &paho.Disconnect{
	//	Properties: nil,
	//	ReasonCode: 0,
	//}
	//err := s.mqttClient.Disconnect(disconnStruct)
	//if err == nil {
	//	s.mqttClient = nil
	//}
	s.LogInfo("Edge Connection Closed\n")
	return nil
}

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *edgeConnectorMQTT) SendStdMessage(msg domain.StdMessageStruct) error {
	topic := s.buildPublishTopicString(msg)
	s.LogDebugf("Sending message for: [%s]  [%+v]", topic, msg)
	s.send(topic, msg)
	return nil
}
//ReadTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorMQTT) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorMQTT) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//ListenForPlcTagChanges implements the interface by subscribing to topics and waiting for related messages
func (s *edgeConnectorMQTT) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	s.LogDebug("BEGIN ListenForEdgeTagChanges")
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
	s.LogDebugf("ListenForPlcTagChanges called for Client %s", clientName)
	//declare the handler for received messages
	//need to subscribe to the topics in the changeFilter
	for key, val := range changeFilter {
		if strings.Contains(key, "EQ") {
			topic := s.buildSubscriptionTopicString(s.tagDataCategory, val)
			err := s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("subscribed to topic %s", topic)
			} else {
				panic(err)
			}
			topic = s.buildSubscriptionTopicString(s.eventCategory, val)
			err = s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("subscribed to topic %s", topic)
			} else {
				panic(err)
			}
		}
	}
}
func (s *edgeConnectorMQTT) send(topic string, message domain.StdMessageStruct) {
	jsonBytes, err := json.Marshal(message)
	retain := false
	if message.Category == "TAGDATA" || message.Category == "EVENT" {
		retain = true
	}
	if err == nil {
		pubStruct := &paho.Publish{
			QoS:        0,
			Retain:     retain,
			Topic:      topic,
			Properties: nil,
			Payload:    jsonBytes,
		}
		pubResp, err := s.mqttConnectionManager.Publish(context.Background(), pubStruct)
		if err != nil {
			s.LogErrorf("mqtt publish error : [%s] / [%+v\n]", err, pubResp)
		} else {
			s.LogInfof("Published to: [%s]", topic)
		}
	} else {
		s.LogErrorf("mqtt publish error : failed to marshal the message %+v [%s]\n", message, err)
	}
}
func (s *edgeConnectorMQTT) buildSubscriptionTopicString(category string, changeFilerVal interface{}) string {
	topic := s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", fmt.Sprintf("%s", changeFilerVal), -1)
	//TODO - more robust and complete template processing
	topic = strings.Replace(topic, "<CATEGORY>", category, -1)
	topic = strings.Replace(topic, "<TAGNAME>", "#", -1)
	return topic
}
func (s *edgeConnectorMQTT) buildPublishTopicString(tag domain.StdMessageStruct) string {
	var topic string = s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", tag.OwningAsset, -1)
	switch tag.Category {
	case "TAGDATA":
		topic = strings.Replace(topic, "<CATEGORY>", s.tagDataCategory, -1)
	case "EVENT":
		topic = strings.Replace(topic, "<CATEGORY>", s.eventCategory, -1)
	default:
		topic = strings.Replace(topic, "<CATEGORY>", "EdgeMessage", -1)
	}
	topic = strings.Replace(topic, "<TAGNAME>", tag.ItemName, -1)
	//TODO - more robust and complete template processing
	return topic
}
func (s *edgeConnectorMQTT) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - how to get history via MQTT - seems like it will depend on what is publishing the MQTT messages
	return []domain.StdMessageStruct{}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *edgeConnectorMQTT) SubscribeToTopic(topic string) error {
	subPropsStruct := &paho.SubscribeProperties{
		SubscriptionIdentifier: nil,
		User:                   nil,
	}
	var subMap = make(map[string]paho.SubscribeOptions)
	subMap[topic] = paho.SubscribeOptions{
		QoS:               0,
		RetainHandling:    0,
		NoLocal:           false,
		RetainAsPublished: false,
	}
	subStruct := &paho.Subscribe{
		Properties:    subPropsStruct,
		Subscriptions: subMap,
	}
	_, err := s.mqttConnectionManager.Subscribe(context.Background(), subStruct)
	if err != nil {
		s.LogErrorf("mqtt subscribe error : %s\n", err)
	} else {
		s.LogInfof("mqtt subscribed to : %s\n", topic)
	}
	return err
}

func (s *edgeConnectorMQTT) tagChangeHandler(m *paho.Publish) {

	var tagStruct domain.StdMessageStruct
	err := json.Unmarshal(m.Payload, &tagStruct)
	if err == nil {
		if s.singleChannel == nil {
			s.ChangeChannels[tagStruct.OwningAsset] <- tagStruct
		} else {
			s.singleChannel <- tagStruct
		}
	} else {
		s.LogErrorf("Failed to unmarchal the payload of the incoming message: %s [%s]", m.Payload, err)
	}
}
