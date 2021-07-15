package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.golang/paho"
	"net"
	"strconv"
	"strings"
	"time"
)

type libreConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttClient *mqtt.Client

	topicTemplate   string
	tagDataCategory string
	eventCategory   string
}

func NewLibreConnectorMQTT(configCategoryName string) *libreConnectorMQTT {
	s := libreConnectorMQTT{
		mqttClient: nil,
	}
	s.SetConfigCategory(configCategoryName)
	s.SetLoggerConfigHook("LIBRCONN")
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
	var conn net.Conn
	var useTlsStr string
	var useTls bool
	var connAck *mqtt.Connack
	var err error
	var server, user, pwd, svcName string
	if server, err = s.GetConfigItem("MQTT_SERVER"); err == nil {
		if useTlsStr, err = s.GetConfigItemWithDefault("MQTT_USE_TLS", "false"); err == nil {
			if pwd, err = s.GetConfigItem("MQTT_PWD"); err == nil {
				if user, err = s.GetConfigItem("MQTT_USER"); err == nil {
					svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
				}
			}
		}
	}
	if err != nil {
		panic("Failed to find configuration data for MQTT connection")
	}

	useTls, err = strconv.ParseBool(useTlsStr)
	if err != nil {
		panic(fmt.Sprintf("Bad value for MQTT_USE-SSL in configuration for libreConnectorMQTT: %s", useTlsStr))
	}
	if useTls {
		conn, err = tls.Dial("tcp", server, nil)
	} else {
		conn, err = net.Dial("tcp", server)
	}
	//conn, err = net.Dial("tcp", server)
	if err != nil {

		s.LogErrorf("Failed to connect to %s: %s", server, err)
		return err
	}

	client := mqtt.NewClient()
	client.Conn = conn

	connStruct := &mqtt.Connect{
		KeepAlive:  300,
		ClientID:   fmt.Sprintf("%v", svcName),
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
		s.LogErrorf("Connect return err=%s", err)
	}
	if connAck.ReasonCode != 0 {
		var cid string
		if s.mqttClient == nil {
			cid = "nil clientid"
		} else {
			cid = s.mqttClient.ClientID
		}
		msg := fmt.Sprintf("%s Failed to connect to %s : %d - %s\n", cid, server, connAck.ReasonCode, connAck.Properties.ReasonString)
		s.LogError("Plc", msg)
	} else {
		s.mqttClient = client
		s.LogInfof("%s Connected to %s\n", s.mqttClient.ClientID, server)
	}
	return err
}

//Close implements the interface by closing the MQTT client
func (s *libreConnectorMQTT) Close() error {
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
	s.LogInfo("Libre Connection Closed\n")
	return err
}

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *libreConnectorMQTT) SendStdMessage(msg domain.StdMessageStruct) error {
	topic := s.buildTopicString(msg)
	s.LogInfof("Sending message for: %+v as %s=>%s", msg, topic, msg.ItemValue)
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
		s.LogErrorf("%s mqtt subscribe error :%s\n", s.mqttClient.ClientID, err)
	} else {
		s.LogInfof("%s mqtt subscribed to : %s\n", s.mqttClient.ClientID, topic)
	}
}

func (s *libreConnectorMQTT) send(topic string, message domain.StdMessageStruct) {
	jsonBytes, err := json.Marshal(message)
	if err == nil {
		pubStruct := &mqtt.Publish{
			QoS:        0,
			Retain:     false,
			Topic:      topic,
			Properties: nil,
			Payload:    jsonBytes,
		}
		pubResp, err := s.mqttClient.Publish(context.Background(), pubStruct)
		if err != nil {
			s.LogErrorf("mqtt publish error : %s / %+v\n", err, pubResp)
		} else {
			s.LogInfof("Published: %s to %s\n", message, topic)
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
