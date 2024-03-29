package services

import (
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
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

func (s *edgeConnectorService) Connect(clientId string) error {
	return s.port.Connect(clientId)
}
func (s *edgeConnectorService) Close() error {
	return s.port.Close()
}
func (s *edgeConnectorService) SendStdMessage(msg domain.StdMessageStruct) error {
	return s.port.SendStdMessage(msg)
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

func (s *edgeConnectorService) StopListeningForTagChanges(client string) error {
	return s.port.StopListeningForTagChanges(client)
}
