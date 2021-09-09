package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	//"os"

	"regexp"
	"strconv"
	"strings"
	"time"
)

type edgeConnectorMQTTv3 struct {
	//inherit logging functions
	libreLogger.LoggingEnabler

	//inherit config functions
	libreConfig.ConfigurationEnabler

	mqttClient     *mqtt.Client
	ChangeChannels map[string]chan domain.StdMessageStruct
	singleChannel  chan domain.StdMessageStruct
	config         map[string]string

	topicTemplate    string
	topicParseRegExp *regexp.Regexp
	tagDataCategory  string
	eventCategory    string
}

func NewEdgeConnectorMQTTv3(configHook string) *edgeConnectorMQTTv3 {
	s := edgeConnectorMQTTv3{
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
func (s *edgeConnectorMQTTv3) Connect(connInfo map[string]interface{}) error {
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
	s.LogDebug("ServiceName = " + svcName)
	if err != nil {
		panic("edgeConnectorMQTTv3 failed to find configuration data for MQTT connection")
	}

	opts := mqtt.NewClientOptions()
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetOrderMatters(false)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(2 * time.Second)
	useTls, err = strconv.ParseBool(useTlsStr)
	if err != nil {
		panic(fmt.Sprintf("Bad value for MQTT_USE_TLS in configuration for PlcConnectorMQTT: %s", useTlsStr))
	}
	if useTls {
		tlsConfig := newTLSConfig()
		opts.AddBroker("ssl://" + server)

		if _, err := s.GetConfigItem("INSECURE_SKIP_VERIFY"); err == nil {
			tlsConfig.InsecureSkipVerify = true
		}

		opts.SetTLSConfig(tlsConfig)
		//conn, err = tls.Dial("tcp", server, nil)
	} else {
		//conn, err = net.Dial("tcp", server)
		opts.AddBroker("tcp://" + server)
	}
	if err != nil {

		s.LogErrorf("Edge", ERROR_MESSAGE_FAILED_TO_CONNECT, server, err)
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
func (s *edgeConnectorMQTTv3) Close() error {
	if s.mqttClient == nil {
		return nil
	}
	client := *s.mqttClient
	client.Disconnect(250)
	time.Sleep(1 * time.Second)
	s.LogInfof("PLC Connection Closed\n")
	return nil
}

//ReadTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorMQTTv3) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = inTagDefs
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *edgeConnectorMQTTv3) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = outTagDefs
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//ListenForPlcTagChanges implements the interface by subscribing to topics and waiting for related messages
func (s *edgeConnectorMQTTv3) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
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
	//s.mqttClient.Router = mqtt.NewSingleHandlerRouter(s.tagChangeHandler)
	//need to subscribe to the topics in the changeFilter
	for key, val := range changeFilter {
		if strings.Contains(key, "EQ") {
			topic := s.buildTopicString(s.tagDataCategory, val)
			err := s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("subscribed to topic %s", topic)
			} else {
				panic(err)
			}
			topic = s.buildTopicString(s.eventCategory, val)
			err = s.SubscribeToTopic(topic)
			if err == nil {
				s.LogInfof("subscribed to topic %s", topic)
			} else {
				panic(err)
			}
		}
	}
}

func (s *edgeConnectorMQTTv3) buildTopicString(category string, changeFilerVal interface{}) string {
	var topic string = s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", fmt.Sprintf("%s", changeFilerVal), -1)
	//TODO - more robust and complete template processing
	topic = strings.Replace(topic, "<CATEGORY>", category, -1)
	topic = strings.Replace(topic, "<TAGNAME>", "#", -1)
	return topic
}
func (s *edgeConnectorMQTTv3) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - how to get history via MQTT - seems like it will depend on what is publishing the MQTT messages
	return []domain.StdMessageStruct{}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *edgeConnectorMQTTv3) SubscribeToTopic(topic string) error {
	c := *s.mqttClient
	if token := c.Subscribe(topic, 0, s.tagChangeHandler); token.Wait() && token.Error() != nil {
		s.LogError(token.Error())
	}
	s.LogDebug("subscribed to " + topic)
	return nil
}

func (s *edgeConnectorMQTTv3) tagChangeHandler(client mqtt.Client, m mqtt.Message) {
	s.LogDebug("BEGIN tagChangeHandler")

	var tagStruct domain.StdMessageStruct
	err := json.Unmarshal(m.Payload(), &tagStruct)
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
