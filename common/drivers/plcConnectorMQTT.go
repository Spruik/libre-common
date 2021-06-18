package drivers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.golang/paho"
	"net"
	"regexp"
	"strings"
	"time"
)

type plcConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttClient     *mqtt.Client
	ChangeChannels map[string]chan domain.StdMessageStruct

	topicTemplate    string
	tagDataCategory  string
	adminCategory    string
	topicParseRegExp *regexp.Regexp
}

func NewPlcConnectorMQTT(configCategoryName string) *plcConnectorMQTT {
	s := plcConnectorMQTT{
		mqttClient:     nil,
		ChangeChannels: make(map[string]chan domain.StdMessageStruct),
	}
	s.SetConfigCategory(configCategoryName)
	s.SetLoggerConfigHook("PlcConnectorMQTT")
	s.tagDataCategory, _ = s.GetConfigItemWithDefault("TAG_DATA_CATEGORY", "Status")
	s.adminCategory, _ = s.GetConfigItemWithDefault("ADMIN_CATEGORY", "Admin")
	s.topicTemplate, _ = s.GetConfigItemWithDefault("TOPIC_TEMPLATE", "<EQNAME>/<CATEGORY>/<TAGNAME>")
	topicRE := s.topicTemplate
	topicRE = strings.Replace(topicRE, "<EQNAME>", "(?P<EQNAME>[A-Za-z0-9_]*)", -1)
	topicRE = strings.Replace(topicRE, "<TAGNAME>", "(?P<TAGNAME>[A-Za-z0-9_]*)", -1)
	topicRE = strings.Replace(topicRE, "<CATEGORY>", fmt.Sprintf("(?P<CATEGORY>%s|%s)", s.adminCategory, s.tagDataCategory), -1)
	s.topicParseRegExp = regexp.MustCompile(topicRE)
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *plcConnectorMQTT) Connect() error {
	var conn net.Conn
	var connAck *mqtt.Connack
	var err error
	var server, user, pwd, svcName string
	if server, err = s.GetConfigItem("MQTT_SERVER"); err == nil {
		if user, err = s.GetConfigItem("MQTT_USER"); err == nil {
			if pwd, err = s.GetConfigItem("MQTT_PWD"); err == nil {
				svcName, err = s.GetConfigItem("MQTT_SVC_NAME")
			}
		}
	}
	if err != nil {
		panic("Failed to find configuration data for libreConnectorMQTT")
	}

	conn, err = net.Dial("tcp", server)
	if err != nil {

		s.LogErrorf("Plc", "Failed to connect to %s: %s", server, err)
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
		return err
	}
	if connAck.ReasonCode != 0 {
		msg := fmt.Sprintf("%s Failed to connect to %s : %d - %s\n", s.mqttClient.ClientID, server, connAck.ReasonCode, connAck.Properties.ReasonString)
		s.LogError("Plc", msg)
		return errors.New(msg)
	} else {
		s.mqttClient = client
		s.LogInfof("%s Connected to %s\n", s.mqttClient.ClientID, server)
	}
	return nil
}

//Close implements the interface by closing the MQTT client
func (s *plcConnectorMQTT) Close() error {
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
	s.LogInfof("%s Connection Closed\n", s.mqttClient.ClientID)
	return err
}

//ReadTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *plcConnectorMQTT) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = inTagDefs
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *plcConnectorMQTT) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = outTagDefs
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//ListenForPlcTagChanges implements the interface by subscribing to topics and waiting for related messages
func (s *plcConnectorMQTT) ListenForPlcTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	clientName := fmt.Sprintf("%s", changeFilter["Client"])
	s.LogDebugf("ListenForPlcTagChanges called for Client %s", clientName)
	s.ChangeChannels[clientName] = c
	//declare the handler for received messages
	s.mqttClient.Router = mqtt.NewSingleHandlerRouter(s.receivedMessageHandler)
	//need to subscribe to the topics in the changeFilter
	var topicSet = make(map[string]struct{})
	for key, val := range changeFilter {
		s.LogDebugf("topic map item: %s=%s", key, val)
		if strings.Contains(key, "Topic") {
			topicSet[s.buildTagTopicString(clientName, val)] = struct{}{}
		}
	}
	for key := range topicSet {
		s.LogDebugf("subscription topic: %s", key)
		s.SubscribeToTopic(fmt.Sprintf("%v", key))
	}
	//also subscribe to the Admin generically
	s.SubscribeToTopic(s.buildAdminTopicString(clientName))
}

func (s *plcConnectorMQTT) buildTagTopicString(eqname string, propname interface{}) string {
	var topic string = s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", eqname, -1)
	topic = strings.Replace(topic, "<CATEGORY>", s.tagDataCategory, -1)
	topic = strings.Replace(topic, "<TAGNAME>", fmt.Sprintf("%s", propname), -1)
	//TODO - more robust and complete template processing?
	return topic
}

func (s *plcConnectorMQTT) buildAdminTopicString(eqname string) string {
	var topic string = s.topicTemplate
	topic = strings.Replace(topic, "<EQNAME>", eqname, -1)
	topic = strings.Replace(topic, "<CATEGORY>", s.adminCategory, -1)
	topic = strings.Replace(topic, "<TAGNAME>", "#", -1)
	//TODO - more robust and complete template processing?
	return topic
}

func (s *plcConnectorMQTT) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = startTS
	_ = endTS
	_ = inTagDefs
	//TODO - how to get history via MQTT - seems like it will depend on what is publishing the MQTT messages
	return []domain.StdMessageStruct{}
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// support functions
//
func (s *plcConnectorMQTT) SubscribeToTopic(topic string) {
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
}

//func (s *plcConnectorMQTT) receivedMessageHandler(m *mqtt.Publish) {
//	eqName, tagName := parseTopic(m.Topic)
//	tagStruct := domain.StdMessageStruct{
//		OwningAsset: eqName,
//		ItemName:     tagName,
//		ItemValue:    string(m.Payload),
//		TagQuality:  128,
//		Err:         nil,
//	}
//	s.ChangeChannels[eqName] <- tagStruct
//}

//func parseTopic(topic string) (string, string) {
//	var eqName, tagName string
//	sndx := strings.Index(topic, "/Status/")
//	if sndx >= 0 {
//		eqName = topic[0:sndx]
//		tagName = topic[sndx+8:]
//	} else {
//		andx := strings.Index(topic, "/Admin/")
//		if andx >= 0 {
//			eqName = topic[0:andx]
//			tagName = topic[sndx+7:]
//		}
//	}
//	return eqName, tagName
//}
func (s *plcConnectorMQTT) receivedMessageHandler(m *mqtt.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")
	tokenMap := s.parseTopic(m.Topic)
	tagStruct := domain.StdMessageStruct{
		OwningAsset: tokenMap["EQNAME"],
		ItemName:    tokenMap["TAGNAME"],
		ItemValue:   string(m.Payload),
		TagQuality:  128,
		Err:         nil,
		Category:    tokenMap["CATEGORY"],
	}
	s.ChangeChannels[tokenMap["EQNAME"]] <- tagStruct
}

func (s *plcConnectorMQTT) parseTopic(topic string) map[string]string {
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
