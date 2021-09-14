package drivers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/DrmagicE/gmqtt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

// EdgeConnectorConfig is a struct to re-create the LibreLogger configuration for edgeConnectorMQTT on the fly for various test cases
type EdgeConnectorConfig struct {
	LoggerHook         string `json:"loggerHook"`
	MqttServer         string `json:"MQTT_SERVER"`
	MqttUser           string `json:"MQTT_USER"`
	MqttPassword       string `json:"MQTT_PWD"`
	MqttServiceName    string `json:"MQTT_SVC_NAME"`
	InsecureSkipVerify string `json:"INSECURE_SKIP_VERIFY"`
	TopicTemplate      string `json:"TOPIC_TEMPLATE"`
	TagDataCategory    string `json:"TAG_DATA_CATEGORY"`
	EventCategory      string `json:"EVENT_CATEGORY"`
}

// EdgeConnectorTestConfig is a struct to re-create the LibreLogger configuration for edgeConnectorMQTT on the fly for various test cases
type EdgeConnectorTestConfig struct {
	LibreLogger struct {
		DefaultLevel       string              `json:"defaultLevel"`
		DefaultDestination string              `json:"defaultDestination"`
		Loggers            []map[string]Logger `json:"loggers"`
	} `json:"libreLogger"`
	EdgeConnectorConfig EdgeConnectorConfig `json:"libreEdgeConnector"`
}

// TestListener is a testcase helper struct that defines what URL and if TLS is used for the MQTT Server that is temporarily spun up
type TestListener struct {
	URL       string
	IsTLS     bool
	TlsConfig *tls.Config
}

// EdgeConnectorMqttTestCase is a test case for testing the EdgeConnectorMqttTestCase
type EdgeConnectorMqttTestCase struct {
	Name                    string
	EdgeConnectorMqttConfig EdgeConnectorTestConfig
	BrokerConfig            TestListener
	IsError                 bool
	IsTimeoutError          bool
	StartMqttServer         bool
}

// EdgeConnectorMqttTestCases is a list of test cases to execute
var EdgeConnectorMqttTestCases = []EdgeConnectorMqttTestCase{
	{
		Name: "TLS Server on 8883 with Certs and same Certs",
		EdgeConnectorMqttConfig: ApplyDefaultConfig(EdgeConnectorTestConfig{
			EdgeConnectorConfig: EdgeConnectorConfig{
				InsecureSkipVerify: "true",
				MqttServer:         "tls://127.0.0.1:8883",
			},
		}),
		BrokerConfig: TestListener{
			URL:       "127.0.0.1:8883",
			IsTLS:     true,
			TlsConfig: GetTlsConfig("server", false),
		},
		IsError:         false,
		IsTimeoutError:  false,
		StartMqttServer: true,
	},
	{
		Name: "TLS Server on 8883 with Skip Verify and different Certs",
		EdgeConnectorMqttConfig: ApplyDefaultConfig(EdgeConnectorTestConfig{
			EdgeConnectorConfig: EdgeConnectorConfig{
				InsecureSkipVerify: "true",
				MqttServer:         "tls://127.0.0.1:8883",
			},
		}),
		BrokerConfig: TestListener{
			URL:       "127.0.0.1:8883",
			IsTLS:     true,
			TlsConfig: GetTlsConfig("server2", false),
		},
		IsError:         false,
		IsTimeoutError:  false,
		StartMqttServer: true,
	},
	{
		Name: "TLS Server on 8883 with different certs - expect failure",
		EdgeConnectorMqttConfig: ApplyDefaultConfig(EdgeConnectorTestConfig{
			EdgeConnectorConfig: EdgeConnectorConfig{
				InsecureSkipVerify: "false",
				MqttServer:         "tls://127.0.0.1:8883",
			},
		}),
		BrokerConfig: TestListener{
			URL:       "127.0.0.1:8883",
			IsTLS:     true,
			TlsConfig: GetTlsConfig("server2", false),
		},
		IsError:         true,
		StartMqttServer: true,
		IsTimeoutError:  true,
	},
}

func TestEdgeConnectorMQTT(t *testing.T) {

	for _, testCase := range EdgeConnectorMqttTestCases {
		t.Logf("INFO | %s executing testcase", testCase.Name)
		thisBrokerRunning := false
		// Create a temp config file `test.json`
		file, _ := json.MarshalIndent(testCase.EdgeConnectorMqttConfig, "", " ")
		_ = ioutil.WriteFile("test.json", file, 0644)

		// Initialize standard Libre Libraries - use our temp file
		libreConfig.Initialize("test.json")
		libreLogger.Initialize("libreLogger")

		// Create an EdgeConnectorMQTT
		edgeConnectorDriver := NewEdgeConnectorMQTT("libreEdgeConnector")

		// Handle Broker Commands
		runBroker := make(chan bool)

		if testCase.StartMqttServer {
			ctx, _ := context.WithCancel(context.Background())
			// Create a listener
			var ln net.Listener
			var err error
			if testCase.BrokerConfig.IsTLS {
				ln, err = tls.Listen("tcp", testCase.BrokerConfig.URL, testCase.BrokerConfig.TlsConfig)
				if err != nil {
					t.Errorf("failed to setup test case %s, expected no error; got %s", testCase.Name, err)
					t.Fail()
				}
			} else {
				ln, err = net.Listen("tcp", testCase.BrokerConfig.URL)
				if err != nil {
					t.Errorf("failed to setup test case %s, expected no error; got %s", testCase.Name, err)
					t.Fail()
				}
			}

			// Initialize an MQTT Server
			s := server.New(
				server.WithTCPListener(ln),
			)
			var subService server.SubscriptionService
			err = s.Init(server.WithHook(server.Hooks{
				OnConnected: func(ctx context.Context, client server.Client) {
					// add subscription for a client when it is connected
					subService.Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
						TopicFilter: "topic",
						QoS:         packets.Qos0,
					})
				},
			}))
			if err != nil {
				t.Errorf("failed to setup test case %s, expected no error; got %s", testCase.Name, err)
				t.Fail()
			}

			subService = s.SubscriptionService()
			_ = s.RetainedService()
			_ = s.Publisher()

			// Listen to Commands
			breakout := false
			go func() {
				defer s.Stop(ctx)
				t.Logf("INFO | %s Listenning to broker commands", testCase.Name)
				// Start Broker
				go func() {
					thisBrokerRunning = true
					err = s.Run()
					thisBrokerRunning = false
					if err != nil {
						panic(err)
					}
					fmt.Printf("%s broker starter complete\n", testCase.Name)
				}()
				for !breakout {
					switch <-runBroker {
					case false:
						t.Logf("INFO | %s Stop Broker", testCase.Name)
						if s != nil {
							s.Stop(ctx)
						}
						breakout = true
					}
				}
				fmt.Printf("%s broker stopper complete\n", testCase.Name)
			}()
		}
		errChan := make(chan error)

		go func() {
			errChan <- edgeConnectorDriver.Connect("test")
		}()

		select {
		case err := <-errChan:
			if testCase.IsError && err == nil {
				t.Errorf("test TestEdgeConnectorMQTT %s expected to failed but didn't", testCase.Name)
				t.Fail()
			} else if !testCase.IsError && err != nil {
				t.Errorf("test TestEdgeConnectorMQTT %s didn't expect to fail; got %s", testCase.Name, err)
				t.Fail()
			}
		case <-time.After(time.Second * 20):
			edgeConnectorDriver.Close()
			if !testCase.IsTimeoutError && thisBrokerRunning && edgeConnectorDriver != nil {
				t.Errorf("test case %s failed due to connection timeout", testCase.Name)
				t.Fail()
			}
		}

		// Do anything else with the EdgeConnector Here
		//
		//

		// Close EdgeConnector
		edgeConnectorDriver.Close()

		// Stop Broker (if running)
		runBroker <- false

		defer os.Remove("test.json")
	}
	defer os.Remove("test.json")

}

// ApplyDefaultConfig just sets anything requried to a sensible default
func ApplyDefaultConfig(parseConfig EdgeConnectorTestConfig) EdgeConnectorTestConfig {
	parseConfig.LibreLogger.DefaultLevel = "DEBUG"
	parseConfig.LibreLogger.DefaultDestination = "CONSOLE"
	parseConfig.LibreLogger.Loggers = make([]map[string]Logger, 0)

	temp := make(map[string]Logger)
	temp["libreEdgeConnector"] = Logger{
		Topic: "EDGECONN",
		Level: "DEBUG",
	}
	parseConfig.LibreLogger.Loggers = append(parseConfig.LibreLogger.Loggers, temp)

	if parseConfig.EdgeConnectorConfig.LoggerHook == "" {
		parseConfig.EdgeConnectorConfig.LoggerHook = "libreEdgeConnector"
	}

	if parseConfig.EdgeConnectorConfig.MqttServer == "" {
		parseConfig.EdgeConnectorConfig.MqttServer = "mqtt://127.0.0.1:1883"
	}
	if parseConfig.EdgeConnectorConfig.MqttUser == "" {
		parseConfig.EdgeConnectorConfig.MqttUser = "public"
	}
	if parseConfig.EdgeConnectorConfig.MqttPassword == "" {
		parseConfig.EdgeConnectorConfig.MqttPassword = "admin"
	}
	if parseConfig.EdgeConnectorConfig.MqttServiceName == "" {
		parseConfig.EdgeConnectorConfig.MqttServiceName = "edgeConnectorSvc"
	}
	if parseConfig.EdgeConnectorConfig.TopicTemplate == "" {
		parseConfig.EdgeConnectorConfig.TopicTemplate = "<EQNAME>/<CATEGORY>/<TAGNAME>"
	}
	if parseConfig.EdgeConnectorConfig.EventCategory == "" {
		parseConfig.EdgeConnectorConfig.EventCategory = "EdgeEvent"
	}
	if parseConfig.EdgeConnectorConfig.TagDataCategory == "" {
		parseConfig.EdgeConnectorConfig.TagDataCategory = "EdgeTagChange"
	}
	return parseConfig
}
