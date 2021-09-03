package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"time"
)

type graphqlSubscriptionService struct {
	port ports.EdgeConnectorPort
}

func NewGraphqlSubscriptionService(port ports.EdgeConnectorPort) *graphqlSubscriptionService {
	return &graphqlSubscriptionService{
		port: port,
	}
}

var graphqlSubscriptionServiceInstance *graphqlSubscriptionService = nil

func SetGraphqlSubscriptionServiceInstance(inst *graphqlSubscriptionService) {
	graphqlSubscriptionServiceInstance = inst
}
func GetGraphqlSubscriptionServiceInstance() *graphqlSubscriptionService {
	return graphqlSubscriptionServiceInstance
}

func (s *graphqlSubscriptionService) Connect(clientId string) error {
	return s.port.Connect(clientId)
}
func (s *graphqlSubscriptionService) Close() error {
	return s.port.Close()
}
func (s *graphqlSubscriptionService) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.ReadTags(inTagDefs)
}
func (s *graphqlSubscriptionService) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.WriteTags(outTagDefs)
}
func (s *graphqlSubscriptionService) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	s.port.ListenForEdgeTagChanges(c, changeFilter)
}
func (s *graphqlSubscriptionService) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.GetTagHistory(startTS, endTS, inTagDefs)
}
