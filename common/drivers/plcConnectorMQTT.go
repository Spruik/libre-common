package drivers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/drivers/autopaho"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/eclipse/paho.golang/paho"
)

type plcConnectorMQTT struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	mqttConnectionManager *autopaho.ConnectionManager
	mqttClient            *paho.Client
	ChangeChannels        map[string]chan domain.StdMessageStruct

	topicTemplateList    []string
	topicParseRegExpList []*regexp.Regexp

	listenMutex sync.Mutex
}

func NewPlcConnectorMQTT(configHook string) *plcConnectorMQTT {
	s := plcConnectorMQTT{
		mqttClient:           nil,
		ChangeChannels:       make(map[string]chan domain.StdMessageStruct),
		topicTemplateList:    make([]string, 0),
		topicParseRegExpList: make([]*regexp.Regexp, 0),
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
	var err error
	var server, user, pwd, svcName string
	if server, err = s.GetConfigItem("MQTT_SERVER"); err == nil {
		if pwd, err = s.GetConfigItem("MQTT_PWD"); err == nil {
			if user, err = s.GetConfigItem("MQTT_USER"); err == nil {
				svcName, _ = s.GetConfigItem("MQTT_SVC_NAME")
			}
		}
	}
	serverUrl, err := url.Parse(server)
	if err != nil {
		panic("plcConnectorMQTT failed to find configuration data for MQTT connection")
	}
	cliCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{serverUrl},
		KeepAlive:         300,
		ConnectRetryDelay: 10 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			s.LogInfo("mqtt connection up")
		},
		OnConnectError: func(err error) { s.LogError("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: svcName,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				s.receivedMessageHandler(m)
			}),
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
	if skip, err := s.GetConfigItem("INSECURE_SKIP_VERIFY"); err == nil && strings.EqualFold(skip, "true") {
		tlsConfig.InsecureSkipVerify = true
	}
	cliCfg.TlsCfg = &tlsConfig
	cliCfg.Debug = log.New(os.Stdout, "autoPaho", 1)
	cliCfg.PahoDebug = log.New(os.Stdout, "paho", 1)
	cliCfg.SetUsernamePassword(user, []byte(pwd))
	ctx, _ := context.WithCancel(context.Background())
	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		s.LogErrorf("PlcConnector failed initial mqtt connection to %s, expected no error; got %s", cliCfg.BrokerUrls, err)
	}
	err = cm.AwaitConnection(ctx)
	s.mqttConnectionManager = cm
	return err
}

//Close implements the interface by closing the MQTT client
func (s *plcConnectorMQTT) Close() error {
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
	s.LogInfof("PLC Connection Closed\n")
	//return err
	return nil
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
	s.mqttClient.Router = paho.NewSingleHandlerRouter(s.receivedMessageHandler)
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
func (s *plcConnectorMQTT) Unsubscribe(equipmentId *string, topicList []string) error {
	u := paho.Unsubscribe{
		Topics:     topicList,
		Properties: nil,
	}
	_, err := s.mqttConnectionManager.Unsubscribe(context.Background(), &u)
	return err
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
	_, err := s.mqttConnectionManager.Subscribe(context.Background(), subStruct)
	if err != nil {
		s.LogErrorf("mqtt subscribe error : %s\n", err)
	} else {
		s.LogInfof("mqtt subscribed to : %s\n", topic)
	}
}

func (s *plcConnectorMQTT) receivedMessageHandler(m *paho.Publish) {
	s.LogDebug("BEGIN tagChangeHandler")
	tokenMap := s.parseTopic(m.Topic)
	tagStruct := domain.StdMessageStruct{
		OwningAsset:      tokenMap["EQNAME"],
		ItemName:         tokenMap["TAGNAME"],
		ItemValue:        string(m.Payload),
		TagQuality:       128,
		Err:              nil,
		ChangedTimestamp: time.Now(),
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
