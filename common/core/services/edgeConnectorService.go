package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"time"
)

type edgeConnectorService struct {
	port ports.EdgeConnectorPort
}

func NewEdgeConnectorService(port ports.EdgeConnectorPort) *edgeConnectorService {
	return &edgeConnectorService{
		port: port,
	}
}

var edgeConnectorServiceInstance *edgeConnectorService = nil

func SetEdgeConnectorServiceInstance(inst *edgeConnectorService) {
	edgeConnectorServiceInstance = inst
}
func GetEdgeConnectorServiceInstance() *edgeConnectorService {
	return edgeConnectorServiceInstance
}

func (s *edgeConnectorService) Connect(connInfo map[string]interface{}) error {
	return s.port.Connect(connInfo)
}
func (s *edgeConnectorService) Close() error {
	return s.port.Close()
}
func (s *edgeConnectorService) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.ReadTags(inTagDefs)
}
func (s *edgeConnectorService) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.WriteTags(outTagDefs)
}
func (s *edgeConnectorService) ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	s.port.ListenForEdgeTagChanges(c, changeFilter)
}
func (s *edgeConnectorService) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.port.GetTagHistory(startTS, endTS, inTagDefs)
}
