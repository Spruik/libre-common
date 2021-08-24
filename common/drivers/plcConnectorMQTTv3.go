package drivers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	//"os"

	//"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type plcConnectorMQTTv3 struct {
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

func NewPlcConnectorMQTTv3(configCategoryName string) *plcConnectorMQTTv3 {
	s := plcConnectorMQTTv3{
		mqttClient:           nil,
		ChangeChannels:       make(map[string]chan domain.StdMessageStruct),
		topicTemplateList:    make([]string, 0, 0),
		topicParseRegExpList: make([]*regexp.Regexp, 0, 0),
		listenMutex:          sync.Mutex{},
	}
	s.SetConfigCategory(configCategoryName)
	s.SetLoggerConfigHook("PlcConnectorMQTT")
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
func (s *plcConnectorMQTTv3) Connect() error {
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
		panic("Failed to find configuration data for MQTT connection")
	}

	opts := mqtt.NewClientOptions()
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetOrderMatters(false)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(2 * time.Second)
	useTls, err = strconv.ParseBool(useTlsStr)
	if err != nil {
		panic(fmt.Sprintf("Bad value for MQTT_USE-SSL in configuration for PlcConnectorMQTT: %s", useTlsStr))
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
func (s *plcConnectorMQTTv3) Close() error {
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
func (s *plcConnectorMQTTv3) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = inTagDefs
	//TODO - need top figure out what topic/message to publish that will request a read from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//WriteTags implements the interface by generating an MQTT message to the PLC, waiting for the result
func (s *plcConnectorMQTTv3) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	_ = outTagDefs
	//TODO - need top figure out what topic/message to publish that will request a write from the PLC
	//  messaging partner
	return []domain.StdMessageStruct{}
}

//ListenForPlcTagChanges implements the interface by subscribing to topics and waiting for related messages
func (s *plcConnectorMQTTv3) ListenForPlcTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	clientName := fmt.Sprintf("%s", changeFilter["Client"])
	s.LogDebugf("ListenForPlcTagChanges called for Client %s", clientName)
	s.ChangeChannels[clientName] = c
	//declare the handler for received messages
	//s.mqttClient.Router = mqtt.NewSingleHandlerRouter(s.receivedMessageHandler)
	//need to subscribe to the topics in the changeFilter
	var topicSet = make(map[string]struct{})
	for key, val := range changeFilter {
		s.LogDebugf("topic map item: %s=%s", key, val)
		if strings.Contains(key, "Topic") {
			for _, tmpl := range s.topicTemplateList {
				var topic string = tmpl
				topic = strings.Replace(topic, "<EQNAME>", clientName, -1)
				topic = strings.Replace(topic, "<TAGNAME>", fmt.Sprintf("%s", val), -1)
				s.LogDebug(topic)
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
	s.LogDebug("topicSet....................")
	s.LogDebug(topicSet)
	for key := range topicSet {
		s.LogDebugf("subscription topic: %s", key)
		go s.SubscribeToTopic(fmt.Sprintf("%v", key))
	}
}

func (s *plcConnectorMQTTv3) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
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
func (s *plcConnectorMQTTv3) SubscribeToTopic(topic string) {
	c := *s.mqttClient
	if token := c.Subscribe(topic, 0, s.receivedMessageHandler); token.Wait() && token.Error() != nil {
		s.LogError(token.Error())
	}
	s.LogDebug("subscribed to " + topic)
}

func (s *plcConnectorMQTTv3) receivedMessageHandler(client mqtt.Client, msg mqtt.Message) {
	s.LogDebug("BEGIN tagChangeHandler")
	tokenMap := s.parseTopic(msg.Topic())
	tagStruct := domain.StdMessageStruct{
		OwningAsset: tokenMap["EQNAME"],
		ItemName:    tokenMap["TAGNAME"],
		ItemValue:   string(msg.Payload()),
		TagQuality:  128,
		Err:         nil,
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

func (s *plcConnectorMQTTv3) parseTopic(topic string) map[string]string {
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
func newTLSConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile("samplecerts/CAfile.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair("samplecerts/client-crt.pem", "samplecerts/client-key.pem")
	if err != nil {
		log.Println(err)
	}
	if cert.Certificate != nil {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			panic(err)
		}
		fmt.Println(cert.Leaf)
		// Just to print out the client certificate..
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			panic(err)
		}
		fmt.Println(cert.Leaf)
	}
	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}
