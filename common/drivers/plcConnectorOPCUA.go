package drivers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/queries"
	"github.com/Spruik/libre-common/common/core/services"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
)

type plcConnectorOPCUA struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	uaClient           *opcua.Client
	subscription       *opcua.Subscription
	connectionContext  context.Context
	ChangeChannels     map[string]chan domain.StdMessageStruct
	aliasSystem        string
	nodeMap            map[string]uint32
	clientHandleMap    map[uint32]string
	monitoredItemIdMap map[string]uint32
}

func NewPlcConnectorOPCUA(configHook string) *plcConnectorOPCUA {
	s := plcConnectorOPCUA{
		ChangeChannels:     map[string]chan domain.StdMessageStruct{},
		nodeMap:            make(map[string]uint32),
		clientHandleMap:    make(map[uint32]string),
		monitoredItemIdMap: make(map[string]uint32),
	}
	s.SetConfigCategory(configHook)
	loggerHook, cerr := s.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	s.SetLoggerConfigHook(loggerHook)
	s.aliasSystem, _ = s.GetConfigItemWithDefault("aliasSystem", "OPCUA")
	return &s
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// interface functions
//

func (s *plcConnectorOPCUA) Connect() error {
	// check the status of the connection
	// if you cannot connect, retry until you can.
	for {
		if s.uaClient == nil {
			s.connect()
		}
		if s.uaClient != nil {
			switch s.uaClient.State() {
			case 0: // Disconnected
				s.connect()
			case 1: // Connected
				s.LogDebug(s.uaClient.State())
				return nil
			case 2:
				s.LogDebug(s.uaClient.State())
			case 3: // disconnected
				s.LogDebug(s.uaClient.State())
			case 4: // reconnecting
				s.LogDebug(s.uaClient.State())
			default:
				s.LogDebug(s.uaClient.State())
			}
			s.LogDebug(s.uaClient.State())
		} else {
			s.LogDebug("Failed to connect to OPCUA, retrying in 5 seconds")
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *plcConnectorOPCUA) connect() error {
	s.connectionContext = context.Background()
	//Grab server address config
	endpointStr, err := s.GetConfigItem("ENDPOINT")
	if err == nil {
		s.LogDebug("Config found:  ENDPOINT: " + endpointStr)
	} else {
		s.LogError("Config read failed:  ENDPOINT", err)
		panic("plcConnectorOPCUA failed to find configuration data for OPCUA connection")
	}

	s.LogDebug("Retrieving available endpoints from OPC server")
	endpoints, err := opcua.GetEndpoints(context.Background(), endpointStr)
	if err == nil {
		var endPointDesc *ua.EndpointDescription
		for i := 0; i < len(endpoints); i++ {
			endPointDesc = endpoints[i]
			s.LogDebug(">>>>>>>>>>>>>")
			s.LogDebug("Endpoint EndpointURL            : " + endPointDesc.EndpointURL)
			s.LogDebug("Endpoint SecurityLevel          : " + fmt.Sprintf("%v", endPointDesc.SecurityLevel))
			s.LogDebug("Endpoint SecurityMode           : " + endPointDesc.SecurityMode.String())
			s.LogDebug("Endpoint Server.ProductURI      : " + endPointDesc.Server.ProductURI)
			s.LogDebug("Endpoint Server.ApplicationURI  : " + endPointDesc.Server.ApplicationURI)
			s.LogDebug("Endpoint Server.DiscoveryProfileURI: " + endPointDesc.Server.DiscoveryProfileURI)
			s.LogDebug("Endpoint Server.GatewayServerURI: " + endPointDesc.Server.GatewayServerURI)
			s.LogDebug("Endpoint Server.ApplicationName : " + endPointDesc.Server.ApplicationName.Text)
			s.LogDebug("Endpoint TransportProfileURI    : " + endPointDesc.TransportProfileURI)
			s.LogDebug("Endpoint SecurityPolicyURI      : " + endPointDesc.SecurityPolicyURI)
			s.LogDebug("<<<<<<<<<<<<<")
		}

	} else {
		s.LogError("Error retrieving available endpoints from OPC server: ", err)
		return (err)
	}

	endpoint := opcua.SelectEndpoint(endpoints, "None", ua.MessageSecurityModeFromString("None"))
	if endpoint == nil {
		return nil
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy("None"),
		opcua.SecurityModeString("None"),
		opcua.CertificateFile(""),
		opcua.PrivateKeyFile(""),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(endpoint, ua.UserTokenTypeAnonymous),
	}

	override, err := s.GetConfigItem("OVERRIDE_ENDPOINTURL")
	if err == nil {
		s.uaClient = opcua.NewClient(override, opts...)
	} else {
		s.uaClient = opcua.NewClient(endpoint.EndpointURL, opts...)
	}
	err = s.uaClient.Connect(s.connectionContext)
	if err == nil {
		s.LogInfo("OPCUA", "Connected to OPCUA server at: "+endpoint.EndpointURL)
	} else {
		s.LogError("OPCUA", "Failed in OPCUA connect: ", err)
	}

	return err
}
func (s *plcConnectorOPCUA) Close() error {
	if s.uaClient != nil {
		err := s.uaClient.Close()
		if err != nil {
			s.LogError("OPCUA", "Failed in OPCUA close: ", err)
		}
		return err
	}
	return nil
}
func (s *plcConnectorOPCUA) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - use the "read" facility in OPCUA - should be straightforward
	_ = inTagDefs
	return []domain.StdMessageStruct{}
}
func (s *plcConnectorOPCUA) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO - use the "write" facility in OPCUA - should be straightforward
	_ = outTagDefs
	return []domain.StdMessageStruct{}
}
func (s *plcConnectorOPCUA) ListenForPlcTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	//TODO - use the "subscribe" processing to get changes for the given tags - need to figure out things
	//  like how the interval is configured

	clientName := fmt.Sprintf("%s", changeFilter["Client"])
	s.LogDebugf("ListenForPlcTagChanges called for Client %s", clientName)
	s.ChangeChannels[clientName] = c

	notifyCh := make(chan *opcua.PublishNotificationData)

	sub, err := s.uaClient.Subscribe(&opcua.SubscriptionParameters{
		Interval: time.Second,
	}, notifyCh)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created subscription with id %v", sub.SubscriptionID)
	s.subscription = sub
	for key, val := range changeFilter {
		if strings.Index(key, "Topic") == 0 {
			id, err := ua.ParseNodeID(val.(string))
			if err != nil {
				s.LogInfof("Skipping OPCUA subscription to item %s because it's external name %s is not compliant", key, val)
				continue
			}
			clientHandle, ok := s.nodeMap[val.(string)]
			if !ok {
				clientHandle = uint32(len(s.nodeMap) + 1)
				s.nodeMap[val.(string)] = clientHandle
				s.clientHandleMap[clientHandle] = val.(string)
			}
			miCreateRequest := valueRequest(id, clientHandle)

			res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
			if err != nil || res.Results[0].StatusCode != ua.StatusOK {
				s.LogErrorf("failed to subscribe to node %s : %s", val.(string), err)
				temp := err.Error()
				tagData := domain.StdMessageStruct{
					Err:              &temp,
					ItemName:         val.(string),
					ChangedTimestamp: time.Now().UTC(),
					Category:         "TAGDATA",
				}
				s.ChangeChannels[clientName] <- tagData
				continue
			}
			for _, v := range res.Results {
				s.monitoredItemIdMap[val.(string)] = v.MonitoredItemID
			}

		}
	}
	go s.startSubscription(clientName, s.connectionContext, notifyCh)
}
func valueRequest(nodeID *ua.NodeID, handle uint32) *ua.MonitoredItemCreateRequest {
	return opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
}
func (s *plcConnectorOPCUA) Unsubscribe(equipmentId *string, topicList []string) error {
	if s.subscription != nil {
		for _, node := range topicList {
			s.subscription.Unmonitor(s.monitoredItemIdMap[node])
		}
	}
	return nil
}

func (s *plcConnectorOPCUA) buildAliasMapForEquipment(eqName string) map[string]string {
	txn := services.GetLibreDataStoreServiceInstance().BeginTransaction(false, "aliasCheck")
	defer txn.Dispose()
	ret, err := queries.GetAliasPropertyNamesForSystem(txn, s.aliasSystem, eqName)
	if err == nil {
		return ret
	} else {
		panic(err)
	}
}

func (s *plcConnectorOPCUA) getPropNameForAlias(extName string, eqName string) string {
	txn := services.GetLibreDataStoreServiceInstance().BeginTransaction(false, "stdCheck")
	defer txn.Dispose()
	intName, err := queries.GetPropertyNameForSystemAlias(txn, s.aliasSystem, extName, eqName)
	if err == nil {
		return intName
	}
	return extName
}

func (s *plcConnectorOPCUA) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	//TODO
	_ = startTS
	_ = endTS
	_ = inTagDefs
	panic("implement me")
}

//
///////////////////////////////////////////////////////////////////////////////////////////////////////////
// local functions
//
func (s *plcConnectorOPCUA) startSubscription(clientName string, ctx context.Context, notifyCh chan *opcua.PublishNotificationData) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-notifyCh:
			if res.Error != nil {
				log.Print(res.Error)
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					data := item.Value.Value.Value()
					s.LogDebugf("MonitoredItem with client handle %v = %v", item.ClientHandle, data)
					tagData := domain.StdMessageStruct{
						OwningAsset:      "", //will be completed by channel listener
						ItemName:         s.clientHandleMap[item.ClientHandle],
						ItemValue:        fmt.Sprintf("%v", data),
						TagQuality:       128,
						Err:              nil,
						ChangedTimestamp: item.Value.ServerTimestamp, //time.now.utc
					}
					s.ChangeChannels[clientName] <- tagData
				}

			case *ua.EventNotificationList:
				s.LogDebug("recieved an opcua event notification, but we don't handle these yet")

			default:
				s.LogDebugf("what's this publish result? %T", res.Value)
			}
		}
	}
}
