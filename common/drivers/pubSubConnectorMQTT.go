package drivers

import (
	"context"
	"encoding/json"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/drivers/autopaho"
	libreConfig "github.com/Spruik/libre-configuration"
	"github.com/Spruik/libre-logging"
	"github.com/eclipse/paho.golang/paho"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

type pubSubConnectorMQTT struct {
	//inherit logging functions
	libreLogger.LoggingEnabler
	//inherit config functions
	libreConfig.ConfigurationEnabler

	mqttConnectionManager     *autopaho.ConnectionManager
	mqttClient     *paho.Client
	singleChannel  chan *domain.StdMessage
	config         map[string]string
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
	var err error
	var server, user, pwd, svcName string

	//Grab server address config
	server, err = s.GetConfigItem("MQTT_SERVER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SERVER: " + server)
	} else {
		s.LogError("Config read failed:  MQTT_SERVER" , err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab password
	pwd, err = s.GetConfigItem("MQTT_PWD")
	if err == nil {
		s.LogDebug("Config found:  MQTT_PWD: <will not be shown in log>")
	} else {
		s.LogError("Config read failed:  MQTT_PWD" , err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab password
	user, err = s.GetConfigItem("MQTT_USER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_USER: " + user)
	} else {
		s.LogError("Config read failed:  MQTT_USER" , err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab service name
	svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SVC_NAME: " + svcName)
	} else {
		s.LogError("Config read failed:  MQTT_SVC_NAME" , err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
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
			panic("pubSubConnectorMQTT server name specifies unsupported URL scheme")
		}
	} else {
		s.LogError("Server name not valid" , err)
		panic("pubSubConnectorMQTT server name not valid")
	}

	cliCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{MQTTServerURL},
		KeepAlive:         300,
		ConnectRetryDelay: 10 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			s.LogInfo("mqtt connection up")
		},
		OnConnectError: func(err error) { s.LogInfo("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: svcName,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				s.tagChangeHandler(m)
			}),
			OnClientError: func(err error) { s.LogInfo("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					s.LogInfo("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					s.LogInfo("server requested disconnect; reason code: %d\n", d.ReasonCode)
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
func (s *pubSubConnectorMQTT) Close() error {
	s.LogInfo("Edge Connection Closed\n")
	return nil
}

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *pubSubConnectorMQTT) Publish(topic string, payload *json.RawMessage, qos byte, retain bool) error {
	s.LogDebug("Start publishing message to topic "+topic)
	pubStruct := &paho.Publish{
		QoS:        0,
		Retain:     retain,
		Topic:      topic,
		Properties: nil,
		Payload:    *payload,
	}
	pubResp, err := s.mqttConnectionManager.Publish(context.Background(), pubStruct)
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
	//s.mqttClient.Router = paho.NewSingleHandlerRouter(s.tagChangeHandler)
	for _,val := range topicMap{
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
	subPropsStruct := &paho.SubscribeProperties{
		SubscriptionIdentifier: nil,
		User:                   nil,
	}
	var subMap = make(map[string]paho.SubscribeOptions)
	subMap[topic] = paho.SubscribeOptions{
		QoS:               1,
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
		s.LogErrorf(" mqtt subscribe error : %s\n",  err)
	} else {
		s.LogInfof(" mqtt subscribed to : %s\n", topic)
	}
	return err
}

func (s *pubSubConnectorMQTT) tagChangeHandler(m *paho.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")

	message := domain.StdMessage{
		Topic:   m.Topic,
		Payload: (*json.RawMessage)(&m.Payload),
	}
	s.singleChannel <- &message
}
