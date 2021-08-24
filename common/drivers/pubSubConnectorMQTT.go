package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.golang/paho"
)

type pubSubConnectorMQTT struct {
	//inherit logging functions
	libreLogger.LoggingEnabler
	//inherit config functions
	libreConfig.ConfigurationEnabler

	mqttClient    *mqtt.Client
	singleChannel chan *domain.StdMessage
	config        map[string]string
}

func NewPubSubConnectorMQTT() *pubSubConnectorMQTT {
	s := pubSubConnectorMQTT{
		mqttClient: nil,
	}
	s.SetConfigCategory("pubSubConnectorMQTT")
	s.SetLoggerConfigHook("PubSubMQTT")
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *pubSubConnectorMQTT) Connect() error {
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
		panic(fmt.Sprintf("Bad value for MQTT_USE-SSL in configuration for pubSubConnectorMQTT: %s", useTlsStr))
	}
	if useTls {
		if _, err := s.GetConfigItem("INSECURE_SKIP_VERIFY"); err == nil {
			conn, err = tls.Dial("tcp", server, &tls.Config{InsecureSkipVerify: true})
		} else {
			conn, err = tls.Dial("tcp", server, nil)
		}
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
func (s *pubSubConnectorMQTT) Close() error {
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

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *pubSubConnectorMQTT) Publish(topic string, payload *json.RawMessage, qos byte, retain bool) error {
	s.LogDebug("Start publishing message to topic " + topic)
	pubStruct := &mqtt.Publish{
		QoS:        0,
		Retain:     retain,
		Topic:      topic,
		Properties: nil,
		Payload:    *payload,
	}
	pubResp, err := s.mqttClient.Publish(context.Background(), pubStruct)
	if err != nil {
		s.LogErrorf("mqtt publish error : %s / %+v\n", err, pubResp)
	}
	return nil

}
func (s *pubSubConnectorMQTT) Subscribe(c chan *domain.StdMessage, topicMap map[string]string) {
	s.LogDebugf("BEGIN Subscribe")
	// the topic always starts with Libre.<EVENT_TYPE>.<ENTITY>
	// where EVENT_TYPE is event, command or subscription
	// and entity is the root type for the payload. Eg, WorkflowInstance or Task

	s.singleChannel = c
	//declare the handler for received messages
	s.mqttClient.Router = mqtt.NewSingleHandlerRouter(s.tagChangeHandler)
	for _, val := range topicMap {
		err := s.SubscribeToTopic(val)
		if err == nil {
			s.LogInfof("Subscribed to topic %s", val)
		} else {
			panic(err)
		}
	}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *pubSubConnectorMQTT) SubscribeToTopic(topic string) error {
	subPropsStruct := &mqtt.SubscribeProperties{
		SubscriptionIdentifier: nil,
		User:                   nil,
	}
	var subMap = make(map[string]mqtt.SubscribeOptions)
	subMap[topic] = mqtt.SubscribeOptions{
		QoS:               1,
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

func (s *pubSubConnectorMQTT) tagChangeHandler(m *mqtt.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")

	message := domain.StdMessage{
		Topic:   m.Topic,
		Payload: (*json.RawMessage)(&m.Payload),
	}
	s.singleChannel <- &message
}
