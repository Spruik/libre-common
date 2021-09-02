package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/drivers/autopaho"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	paho "github.com/eclipse/paho.golang/paho"
)

type libreConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttConnectionManager     *autopaho.ConnectionManager
	mqttClient     *paho.Client

	topicTemplate   string
	tagDataCategory string
	eventCategory   string
}

func NewLibreConnectorMQTT(configHook string) *libreConnectorMQTT {
	s := libreConnectorMQTT{
		mqttClient: nil,
	}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	s.topicTemplate, _ = s.GetConfigItemWithDefault("TOPIC_TEMPLATE", "<EQNAME>/<CATEGORY>/<TAGNAME>")
	s.tagDataCategory, _ = s.GetConfigItemWithDefault("TAG_DATA_CATEGORY", "EdgeTagChange")
	s.eventCategory, _ = s.GetConfigItemWithDefault("EVENT_CATEGORY", "EdgeEvent")
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *libreConnectorMQTT) Connect() error {
	var err error
	var server, user, pwd, svcName string
	if server, err = s.GetConfigItem("MQTT_SERVER"); err == nil {
		if pwd, err = s.GetConfigItem("MQTT_PWD"); err == nil {
			if user, err = s.GetConfigItem("MQTT_USER"); err == nil {
				svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
			}
		}
	}
	serverUrl,err := url.Parse(server)
	if err != nil {
		panic("libreConnectorMQTT failed to find configuration data for MQTT connection")
	}

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{serverUrl},
		KeepAlive:         300,
		ConnectRetryDelay: 10 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")
		},
		OnConnectError: func(err error) { fmt.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: svcName,
			//TODO: no router?
			OnClientError: func(err error) { fmt.Printf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
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
func (s *libreConnectorMQTT) Close() error {
	if s.mqttClient == nil {
		return nil
	}
	disconnStruct := &paho.Disconnect{
		Properties: nil,
		ReasonCode: 0,
	}
	err := s.mqttClient.Disconnect(disconnStruct)
	if err == nil {
		s.mqttClient = nil
	}
	s.LogInfo("Libre Connection Closed\n")
	return err
}

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *libreConnectorMQTT) SendStdMessage(msg domain.StdMessageStruct) error {
	topic := s.buildTopicString(msg)
	s.LogDebugf("Sending message for: %+v as %s=>%s", msg, topic, msg.ItemValue)
	s.send(topic, msg)
	return nil
}

//ReadTags implements the interface by using the services of the PLC connector
func (s *libreConnectorMQTT) ListenForReadTagsRequest(c chan []domain.StdMessageStruct, readTagDefs []domain.StdMessageStruct) {
	_ = c
	_ = readTagDefs
	//TODO -

}

//WriteTags implements the interface by using the services of the PLC connector
func (s *libreConnectorMQTT) ListenForWriteTagsRequest(c chan []domain.StdMessageStruct, writeTagDefs []domain.StdMessageStruct) {
	_ = c
	_ = writeTagDefs
	//TODO -

}

func (s *libreConnectorMQTT) ListenForGetTagHistoryRequest(c chan []domain.StdMessageStruct, startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) {
	_ = c
	_ = startTS
	_ = endTS
	_ = inTagDefs
	//TODO -

}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *libreConnectorMQTT) subscribeToTopic(topic string) {
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
	_, err := s.mqttClient.Subscribe(context.Background(), subStruct)
	if err != nil {
		s.LogErrorf("%s mqtt subscribe error :%s\n", s.mqttClient.ClientID, err)
	} else {
		s.LogInfof("%s mqtt subscribed to : %s\n", s.mqttClient.ClientID, topic)
	}
}

func (s *libreConnectorMQTT) send(topic string, message domain.StdMessageStruct) {
	jsonBytes, err := json.Marshal(message)
	retain := false
	if message.Category == "TAGDATA" {
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
			s.LogErrorf("mqtt publish error : %s / %+v\n", err, pubResp)
		} else {
			s.LogInfof("Published to %s", topic)
		}
	} else {
		s.LogErrorf("mqtt publish error : failed to marshal the message %+v [%s]\n", message, err)
	}
}

func (s *libreConnectorMQTT) buildTopicString(tag domain.StdMessageStruct) string {
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
