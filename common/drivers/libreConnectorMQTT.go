package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/drivers/autopaho"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/eclipse/paho.golang/paho"
)

type libreConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttConnectionManager *autopaho.ConnectionManager
	mqttClient            *paho.Client

	topicTemplate   string
	tagDataCategory string
	eventCategory   string

	ctxCancel context.CancelFunc
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

	//Grab server address config
	server, err = s.GetConfigItem("MQTT_SERVER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SERVER: " + server)
	} else {
		s.LogError("Config read failed:  MQTT_SERVER", err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab password
	pwd, err = s.GetConfigItem("MQTT_PWD")
	if err == nil {
		s.LogDebug("Config found:  MQTT_PWD: <will not be shown in log>")
	} else {
		s.LogError("Config read failed:  MQTT_PWD", err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab user
	user, err = s.GetConfigItem("MQTT_USER")
	if err == nil {
		s.LogDebug("Config found:  MQTT_USER: " + user)
	} else {
		s.LogError("Config read failed:  MQTT_USER", err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	//Grab service name
	svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
	if err == nil {
		s.LogDebug("Config found:  MQTT_SVC_NAME: " + svcName)
	} else {
		s.LogError("Config read failed:  MQTT_SVC_NAME", err)
		panic("pubSubConnectorMQTT failed to find configuration data for MQTT connection")
	}

	MQTTServerURL, err := url.Parse(server)

	if err == nil {
		s.LogDebugf("URL Schema (eg:  mqtt:// ) determines the connection type used by Libre")
		lowerCaseScheme := strings.ToLower(MQTTServerURL.Scheme)
		switch lowerCaseScheme {
		case "mqtt", "tcp":
			s.LogDebugf("URL Scheme: [%s] requires INSECURE connection, Libre edge will not connect if the MQTT broker has TLS enabled", lowerCaseScheme)
		case "ssl", "tls", "mqtts", "mqtt+ssl", "tcps":
			s.LogDebugf("URL Scheme: [%s] requires SECURE connection, TLS is required at the MQTT broker", lowerCaseScheme)
		default:
			s.LogErrorf("URL Scheme: [%s] is not supported", lowerCaseScheme)
			s.LogErrorf("Supported SECURE schema are: ssl, tls, mqtts, mqtt+ssl, tcps")
			s.LogErrorf("Other supported schema are: mqtt, tcp")
			panic("edgeConnectorMQTT server name specifies unsupported URL scheme")
		}
	} else {
		s.LogError("Server name not valid", err)
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
			ClientID:      svcName,
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

	tlsConfig := tls.Config{}
	if skip, tlsErr := s.GetConfigItem("INSECURE_SKIP_VERIFY"); tlsErr == nil && strings.EqualFold(skip, "true") {
		tlsConfig.InsecureSkipVerify = true
	}
	cliCfg.TlsCfg = &tlsConfig
	cliCfg.Debug = log.New(os.Stdout, "autoPaho", 1)
	cliCfg.PahoDebug = log.New(os.Stdout, "paho", 1)
	cliCfg.SetUsernamePassword(user, []byte(pwd))
	ctx, cancel := context.WithCancel(context.Background())
	s.ctxCancel = cancel
	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		s.LogErrorf("LibreConnector failed initial mqtt connection to %s, expected no error; got %s", cliCfg.BrokerUrls, err)
	}
	err = cm.AwaitConnection(ctx)
	s.mqttConnectionManager = cm
	return err
}

//Close implements the interface by closing the MQTT client
func (s *libreConnectorMQTT) Close() error {
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

	if s.ctxCancel != nil {
		s.ctxCancel()
	}

	s.LogInfo("Libre Connection Closed\n")
	return nil
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
		pubResp, publishErr := s.mqttConnectionManager.Publish(context.Background(), pubStruct)
		if publishErr != nil {
			s.LogErrorf("mqtt publish error : %s / %+v\n", publishErr, pubResp)
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
