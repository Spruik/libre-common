package drivers

import (
	"encoding/json"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strconv"
	"strings"
	"time"
)

type libreConnectorMQTTv3 struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttClient *mqtt.Client

	topicTemplate   string
	tagDataCategory string
	eventCategory   string
}
/*
The Libre Connector is currently only configured to Publish. It does not have channels implemented for subscribing.
 */
func NewLibreConnectorMQTTv3(configHook string) *libreConnectorMQTTv3 {
	s := libreConnectorMQTTv3{
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
func (s *libreConnectorMQTTv3) Connect() error {
	var useTlsStr string
	var useTls bool
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

	opts := mqtt.NewClientOptions()
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetDefaultPublishHandler(f)
	useTls, err = strconv.ParseBool(useTlsStr)
	if err != nil {
		panic(fmt.Sprintf("Bad value for MQTT_USE-SSL in configuration for PlcConnectorMQTT: %s", useTlsStr))
	}
	if useTls {
		tlsConfig := newTLSConfig()
		opts.AddBroker("ssl://"+server)
		opts.SetClientID(svcName).SetTLSConfig(tlsConfig)
		//conn, err = tls.Dial("tcp", server, nil)
	} else {
		//conn, err = net.Dial("tcp", server)
		opts.AddBroker("tcp://"+server)
		opts.SetClientID(svcName)
		opts.SetKeepAlive(2 * time.Second)
		opts.SetPingTimeout(1 * time.Second)
	}
	if err != nil {

		s.LogErrorf("Plc", "Failed to connect to %s: %s", server, err)
		return err
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		s.LogError(token.Error())
	}
	s.mqttClient = &client
	reader := client.OptionsReader()
	s.LogInfof("%s Connected to %s\n", reader.ClientID(), server)
	return err
}

//Close implements the interface by closing the MQTT client
func (s *libreConnectorMQTTv3) Close() error {
	if s.mqttClient == nil {
		return nil
	}
	client := *s.mqttClient
	client.Disconnect(250)
	time.Sleep(1 * time.Second)
	s.LogInfof("Libre Connection Closed\n")
	return nil
}

//SendTagChange implements the interface by publishing the tag data to the standard tag change topic
func (s *libreConnectorMQTTv3) SendStdMessage(msg domain.StdMessageStruct) error {
	topic := s.buildTopicString(msg)
	s.LogInfof("Sending message for: %+v as %s=>%s", msg, topic, msg.ItemValue)
	s.send(topic, msg)
	return nil
}

//ReadTags implements the interface by using the services of the PLC connector
func (s *libreConnectorMQTTv3) ListenForReadTagsRequest(c chan []domain.StdMessageStruct, readTagDefs []domain.StdMessageStruct) {
	_ = c
	_ = readTagDefs
	//TODO -

}

//WriteTags implements the interface by using the services of the PLC connector
func (s *libreConnectorMQTTv3) ListenForWriteTagsRequest(c chan []domain.StdMessageStruct, writeTagDefs []domain.StdMessageStruct) {
	_ = c
	_ = writeTagDefs
	//TODO -

}

func (s *libreConnectorMQTTv3) ListenForGetTagHistoryRequest(c chan []domain.StdMessageStruct, startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) {
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
func (s *libreConnectorMQTTv3) SubscribeToTopic(topic string) {
	c:= *s.mqttClient
	if token := c.Subscribe(topic, 0, s.receivedMessageHandler); token.Wait() && token.Error() != nil {
		s.LogError(token.Error())
	}
	s.LogDebug("subscribed to "+topic)
}

func (s *libreConnectorMQTTv3) receivedMessageHandler(client mqtt.Client, msg mqtt.Message) {
	s.LogDebug("RECEIVED MESSAGE on topic:"+msg.Topic())
	//ToDo: Implement receive handler for libreConnector
}
func (s *libreConnectorMQTTv3) send(topic string, message domain.StdMessageStruct) {
	c := *s.mqttClient
	jsonBytes, err := json.Marshal(message)
	if err == nil {
		token := c.Publish(topic,0,false, jsonBytes)
		token.Wait()
		if token.Error() != nil {
			s.LogErrorf("mqtt publish error : %s ", token.Error())
		} else {
			s.LogDebugf("Published: %s to %s\n", message, topic)
		}
	} else {
		s.LogErrorf("mqtt publish error : failed to marshal the message %+v [%s]\n", message, err)
	}
}

func (s *libreConnectorMQTTv3) buildTopicString(tag domain.StdMessageStruct) string {
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
