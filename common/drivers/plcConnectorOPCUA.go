package drivers

import (
	"context"
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
	"log"
	"strings"
	"time"
)

type plcConnectorOPCUA struct {
	//inherit config
	libreConfig.ConfigurationEnabler
	//inherit logging
	libreLogger.LoggingEnabler

	uaClient          *opcua.Client
	connectionContext context.Context
	ChangeChannels    map[string]chan domain.StdMessageStruct
}

func NewPlcConnectorOPCUA() *plcConnectorOPCUA {
	s := plcConnectorOPCUA{
		ChangeChannels: map[string]chan domain.StdMessageStruct{},
	}
	s.SetConfigCategory("plcConnectorOPCUA")
	s.SetLoggerConfigHook("OPCUA")
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

	s.uaClient = opcua.NewClient(endpoint.EndpointURL, opts...)
	err = s.uaClient.Connect(s.connectionContext)
	if err == nil {
		s.LogInfof("OPCUA", "Connected to OPCUA server at: %s", endpoint.EndpointURL)
	} else {
		s.LogError("OPCUA", "Failed in OPCUA connect: ", err)
	}
	return err
}
func (s *plcConnectorOPCUA) Close() error {
	err := s.uaClient.Close()
	if err != nil {
		s.LogError("OPCUA", "Failed in OPCUA close: ", err)
	}
	return err
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

	nodeNames := make([]string, 0, 0)
	for key, val := range changeFilter {
		if strings.Index(key, "Topic") == 0 {
			if strings.Index(fmt.Sprintf("%s", val), "ns=") == 0 {
				//node has a good OPCUA name, so subscribe
				nodeNames = append(nodeNames, fmt.Sprintf("%s", val))
			} else {
				s.LogInfof("Skipping OPCUA subscription to item %s because it's name is not compliant", val)
			}
		}
	}
	s.startChanSub(clientName, s.connectionContext, m, time.Second*5, 0, nodeNames...)
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
				tagData := domain.StdMessageStruct{
					OwningAsset: "", //will be completed by channel listener
					ItemName:    msg.NodeID.String(),
					ItemValue:   fmt.Sprintf("%v", msg.Value.Value()),
					TagQuality:  128,
					Err:         nil,
				}
				s.ChangeChannels[clientName] <- tagData
			}
			time.Sleep(lag)
		}
	}
}
