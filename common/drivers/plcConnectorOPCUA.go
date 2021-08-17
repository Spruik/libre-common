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
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
)

type plcConnectorOPCUA struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	uaClient          *opcua.Client
	connectionContext context.Context
	ChangeChannels    map[string]chan domain.StdMessageStruct

	aliasSystem string
}

func NewPlcConnectorOPCUA(configHook string) *plcConnectorOPCUA {
	s := plcConnectorOPCUA{
		ChangeChannels: map[string]chan domain.StdMessageStruct{},
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
	s.connectionContext = context.Background()
	endpointStr, err := s.GetConfigItem("ENDPOINT")
	if err != nil {
		panic("Failed to find ENDPOINT entry in configuration for plcConnectorOPCUA")
	}
	endpoints, err := opcua.GetEndpoints(endpointStr)
	if err != nil {
		log.Fatal(err)
	}

	endpoint := opcua.SelectEndpoint(endpoints, "None", ua.MessageSecurityModeFromString("None"))
	if endpoint == nil {
		panic("Failed to find suitable endpoint")
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
		s.LogInfof("OPCUA", "Connected to OPCUA server at: %s", endpoint.EndpointURL)
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

	m, err := monitor.NewNodeMonitor(s.uaClient)
	if err != nil {
		log.Fatal(err)
	}

	m.SetErrorHandler(func(_ *opcua.Client, sub *monitor.Subscription, err error) {
		log.Printf("error: sub=%d err=%s", sub.SubscriptionID(), err.Error())
	})

	aliasMap := s.buildAliasMapForEquipment(clientName)
	nodeNames := make([]string, 0, 0)
	for key, val := range changeFilter {
		if strings.Index(key, "Topic") == 0 {
			//var extName string = s.getAliasForPropName(fmt.Sprintf("%s", val), clientName)
			var extName string = aliasMap[fmt.Sprintf("%s", val)]
			if strings.Index(fmt.Sprintf("%s", extName), "ns=") == 0 {
				//node has a good OPCUA name, so subscribe
				nodeNames = append(nodeNames, fmt.Sprintf("%s", extName))
			} else {
				s.LogInfof("Skipping OPCUA subscription to item %s because it's external name %s is not compliant", val, extName)
			}
		}
	}
	go s.startChanSub(clientName, s.connectionContext, m, time.Second*5, 0, nodeNames...)
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

//func (s *plcConnectorOPCUA) getAliasForPropName(stdPropName string, eqName string) string {
//	txn := services.GetLibreDataStoreServiceInstance().BeginTransaction(false, "aliasCheck")
//	defer txn.Dispose()
//	extName, err := queries.GetAliasPropertyNameForSystem(txn, s.aliasSystem, stdPropName, eqName)
//	if err == nil {
//		return extName
//	}
//	return stdPropName
//}

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
func (s *plcConnectorOPCUA) startChanSub(clientName string, ctx context.Context, m *monitor.NodeMonitor, interval, lag time.Duration, nodes ...string) {
	ch := make(chan *monitor.DataChangeMessage, 16)
	sub, err := m.ChanSubscribe(ctx, &opcua.SubscriptionParameters{Interval: interval}, ch, nodes...)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s subscribed with id=%d  (%+v)", clientName, sub.SubscriptionID(), nodes)

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg.Error != nil {
				log.Printf("%s[channel ] sub=%d error=%s", clientName, sub.SubscriptionID(), msg.Error)
			} else {
				log.Printf("%s[channel ] sub=%d ts=%s node=%s value=%v", clientName, sub.SubscriptionID(), msg.SourceTimestamp.UTC().Format(time.RFC3339), msg.NodeID, msg.Value.Value())
				intName := s.getPropNameForAlias(msg.NodeID.String(), clientName)
				tagData := domain.StdMessageStruct{
					OwningAsset: "", //will be completed by channel listener
					ItemName:    intName,
					ItemValue:   fmt.Sprintf("%v", msg.Value.Value()),
					TagQuality:  128,
					Err:         nil,
					ChangedTime: msg.ServerTimestamp,
				}
				s.ChangeChannels[clientName] <- tagData
			}
			time.Sleep(lag)
		}
	}
}
