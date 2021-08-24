package drivers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.golang/paho"
)

type plcConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttClient     *mqtt.Client
	ChangeChannels map[string]chan domain.StdMessageStruct

	topicTemplateList    []string
	topicParseRegExpList []*regexp.Regexp

	listenMutex sync.Mutex
}

func NewPlcConnectorMQTT(configHook string) *plcConnectorMQTT {
	s := plcConnectorMQTT{
		mqttClient:           nil,
		ChangeChannels:       make(map[string]chan domain.StdMessageStruct),
		topicTemplateList:    make([]string, 0, 0),
		topicParseRegExpList: make([]*regexp.Regexp, 0, 0),
		listenMutex:          sync.Mutex{},
	}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	tmplStanza, err := s.GetConfigStanza("TOPIC_TEMPLATES")
	if err == nil {
		for _, child := range tmplStanza.Children {
			s.topicTemplateList = append(s.topicTemplateList, child.Value)
			topicRE := "^" + child.Value + "$"
			topicRE = strings.Replace(topicRE, "<EQNAME>", "(?P{{{EQNAME}}}[A-Za-z0-9_\\/\\-]*)", -1)
			topicRE = strings.Replace(topicRE, "<", "(?P<", -1)
			topicRE = strings.Replace(topicRE, ">", ">[A-Za-z0-9_]*)", -1)
			topicRE = strings.Replace(topicRE, "{{{EQNAME}}}", "<EQNAME>", -1)
			s.topicParseRegExpList = append(s.topicParseRegExpList, regexp.MustCompile(topicRE))
		}
	}

	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//
//Connect implements the interface by creating an MQTT client
func (s *plcConnectorMQTT) Connect() error {
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
		panic(fmt.Sprintf("Bad value for MQTT_USE-SSL in configuration for PlcConnectorMQTT: %s", useTlsStr))
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
	s.LogInfof("PLC Connection Closed\n")
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
			for _, tmpl := range s.topicTemplateList {
				var topic string = tmpl
				topic = strings.Replace(topic, "<EQNAME>", clientName, -1)
				topic = strings.Replace(topic, "<TAGNAME>", fmt.Sprintf("%s", val), -1)
				var i, j int
				i = strings.Index(topic, "<")
				for i >= 0 {
					j = strings.Index(topic, ">")
					topic = topic[0:i] + "+" + topic[j+1:]
					i = strings.Index(topic, "<")
				}
				topicSet[topic] = struct{}{}
			}
		}
	}
	for key := range topicSet {
		s.LogDebugf("subscription topic: %s", key)
		s.SubscribeToTopic(fmt.Sprintf("%v", key))
	}
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

func (s *plcConnectorMQTT) receivedMessageHandler(m *mqtt.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")
	tokenMap := s.parseTopic(m.Topic)
	tagStruct := domain.StdMessageStruct{
		OwningAsset: tokenMap["EQNAME"],
		ItemName:    tokenMap["TAGNAME"],
		ItemValue:   string(m.Payload),
		TagQuality:  128,
		Err:         nil,
		ChangedTime: time.Now(),
		//Category:    tokenMap["CATEGORY"],

	}
	if tagStruct.ItemName == "" {
		tagStruct.ItemNameExt = make(map[string]string)
		for key, val := range tokenMap {
			if key != "EQNAME" && key != "TAGNAME" {
				tagStruct.ItemNameExt[key] = val
			}
		}
	}
	s.ChangeChannels[tokenMap["EQNAME"]] <- tagStruct
}

func (s *plcConnectorMQTT) parseTopic(topic string) map[string]string {
	ret := map[string]string{}
	for _, re := range s.topicParseRegExpList {
		if re.MatchString(topic) {
			matches := re.FindStringSubmatch(topic)
			names := re.SubexpNames()
			for i, name := range names {
				if i > 0 && i <= len(matches) {
					ret[name] = matches[i]
				}
			}
			break
		}
	}
	return ret
}
