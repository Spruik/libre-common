package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.golang/paho"
	"net"
	"regexp"
	"strings"
	"time"
)

type edgeConnectorMQTT struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	mqttClient    *mqtt.Client
	ChangeChannel chan domain.StdMessageStruct
	config        map[string]string

	topicTemplate    string
	topicParseRegExp *regexp.Regexp
	tagDataCategory  string
	eventCategory    string
}

func NewEdgeConnectorMQTT() *edgeConnectorMQTT {
	s := edgeConnectorMQTT{
		mqttClient: nil,
	}
	s.SetConfigCategory("edgeConnectorMQTT")
	s.SetLoggerConfigHook("EDGEMQTT")
	s.topicTemplate, _ = s.GetConfigItemWithDefault("TOPIC_TEMPLATE", "<EQNAME>/Report/<TAGNAME>")
	s.tagDataCategory, _ = s.GetConfigItemWithDefault("TAG_DATA_CATEGORY", "EdgeTagChange")
	s.eventCategory, _ = s.GetConfigItemWithDefault("EVENT_CATEGORY", "EdgeEvent")
	topicRE := s.topicTemplate
	topicRE = strings.Replace(topicRE, "<EQNAME>", "(?P<EQNAME>[A-Za-z0-9_]*)", -1)
	topicRE = strings.Replace(topicRE, "<TAGNAME>", "(?P<TAGNAME>[A-Za-z0-9_]*)", -1)
	s.topicParseRegExp = regexp.MustCompile(topicRE)
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *edgeConnectorMQTT) Connect(connInfo map[string]interface{}) error {
	var conn net.Conn
	var connAck *mqtt.Connack
	var err error

	server, err := s.GetConfigItem("MQTT_SERVER")
	conn, err = net.Dial("tcp", server)
	if err != nil {

		s.LogErrorf("Failed to connect to %s: %s", server, err)
		return err
	}

	client := mqtt.NewClient()
	client.Conn = conn

	user, err := s.GetConfigItem("MQTT_USER")
	pwd, err := s.GetConfigItem("MQTT_PWD")
	svcName, err := s.GetConfigItem("MQTT_SVC_NAME")

	connStruct := &mqtt.Connect{
		KeepAlive:  30,
		ClientID:   svcName,
		CleanStart: true,
		Username:   user,
		Password:   []byte(pwd),
	}

	if user != "" {
		connStruct.UsernameFlag = true
	}
	if pwd != "" {
		connStruct.PasswordFlag = true
	}

	connAck, err = client.Connect(context.Background(), connStruct)
	if err != nil {
		return err
	}
	if connAck.ReasonCode != 0 {
		msg := fmt.Sprintf("%s Failed to connect to %s : %d - %s\n", s.mqttClient.ClientID, server, connAck.ReasonCode, connAck.Properties.ReasonString)
		s.LogError(msg)
		return errors.New(msg)
	} else {
		s.mqttClient = client
		s.LogInfof("%s Connected to %s\n", s.mqttClient.ClientID, server)
	}
	return nil
}

//Close implements the interface by closing the MQTT client
func (s *edgeConnectorMQTT) Close() error {
	if s.mqttClient == nil {
		return nil
	}
	disconnStruct := &mqtt.Disconnect{
		Properties: nil,
		ReasonCode: 0,
	}
	err := s.mqttClient.Disconnect(disconnStruct)
	if err == nil {
		s.mqttClient = nil
	}
	s.LogInfo("Edge Connection Closed\n")
	return err
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
	s.ChangeChannel = c
	//declare the handler for received messages
	s.mqttClient.Router = mqtt.NewSingleHandlerRouter(s.tagChangeHandler)
	//need to subscribe to the topics in the changeFilter
	for key, val := range changeFilter {
		if strings.Contains(key, "EQ") {
			topic := s.buildTopicString(s.tagDataCategory, val)
			err := s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("%s subscribed to topic %s", s.mqttClient.ClientID, topic)
			} else {
				panic(err)
			}
			topic = s.buildTopicString(s.eventCategory, val)
			err = s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("%s subscribed to topic %s", s.mqttClient.ClientID, topic)
			} else {
				panic(err)
			}
		}
	}
}

func (s *edgeConnectorMQTT) buildTopicString(category string, changeFilerVal interface{}) string {
	var topic string = s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", fmt.Sprintf("%s", changeFilerVal), -1)
	//TODO - more robust and complete template processing
	topic = strings.Replace(topic, "<CATEGORY>", category, -1)
	topic = strings.Replace(topic, "<TAGNAME>", "#", -1)
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
	subPropsStruct := &mqtt.SubscribeProperties{
		SubscriptionIdentifier: nil,
		User:                   nil,
	}
	var subMap = make(map[string]mqtt.SubscribeOptions)
	subMap[topic] = mqtt.SubscribeOptions{
		QoS:               0,
		RetainHandling:    0,
		NoLocal:           false,
		RetainAsPublished: false,
	}
	subStruct := &mqtt.Subscribe{
		Properties:    subPropsStruct,
		Subscriptions: subMap,
	}
	_, err := s.mqttClient.Subscribe(context.Background(), subStruct)
	if err != nil {
		s.LogErrorf("%s mqtt subscribe error : %s\n", s.mqttClient.ClientID, err)
	} else {
		s.LogInfof("%s mqtt subscribed to : %s\n", s.mqttClient.ClientID, topic)
	}
	return err
}

func (s *edgeConnectorMQTT) tagChangeHandler(m *mqtt.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")

	var tagStruct domain.StdMessageStruct
	err := json.Unmarshal(m.Payload, &tagStruct)
	if err == nil {
		s.ChangeChannel <- tagStruct
	} else {
		s.LogErrorf("Failed to unmarchal the payload of the incoming message: %s [%s]", m.Payload, err)
	}
	//tokenMap := s.parseTopic(m.Topic)
	//tagStruct := domain.StdMessageStruct{
	//	OwningAsset: tokenMap["EQNAME"],
	//	ItemName:    tokenMap["TAGNAME"],
	//	ItemValue:   string(m.Payload),
	//	TagQuality:  128,
	//	Err:         nil,
	//}
	//s.ChangeChannel <- tagStruct
}

func (s *edgeConnectorMQTT) parseTopic(topic string) map[string]string {
	ret := map[string]string{}
	names := s.topicParseRegExp.SubexpNames()
	matches := s.topicParseRegExp.FindStringSubmatch(topic)
	for i, name := range names {
		if i > 0 && i <= len(matches) {
			ret[name] = matches[i]
		}
	}
	return ret
}
