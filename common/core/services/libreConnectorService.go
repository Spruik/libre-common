package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"time"
)

type libreConnectorService struct {
	libreConnectorPort ports.LibreConnectorPort
}

func NewLibreConnectorService(port ports.LibreConnectorPort) *libreConnectorService {
	return &libreConnectorService{
		libreConnectorPort: port,
	}
}

var libreConnectorServiceInstance *libreConnectorService = nil

func SetLibreConnectorServiceInstance(inst *libreConnectorService) {
	libreConnectorServiceInstance = inst
}
func GetLibreConnectorServiceInstance() *libreConnectorService {
	return libreConnectorServiceInstance
}

func (s *libreConnectorService) Connect() error {
	return s.libreConnectorPort.Connect()
}
func (s *libreConnectorService) Close() error {
	return s.libreConnectorPort.Close()
}
func (s *libreConnectorService) SendStdMessage(msg domain.StdMessageStruct) error {
	return s.libreConnectorPort.SendStdMessage(msg)
}
func (s *libreConnectorService) ListenForReadTagsRequest(c chan []domain.StdMessageStruct, readTagDefs []domain.StdMessageStruct) {
	s.libreConnectorPort.ListenForReadTagsRequest(c, readTagDefs)
}
func (s *libreConnectorService) ListenForWriteTagsRequest(c chan []domain.StdMessageStruct, writeTagDefs []domain.StdMessageStruct) {
	s.libreConnectorPort.ListenForWriteTagsRequest(c, writeTagDefs)
}
func (s *libreConnectorService) ListenForGetTagHistoryRequest(c chan []domain.StdMessageStruct, startTS time.Time, endTS time.Time, histTagDefs []domain.StdMessageStruct) {
	s.libreConnectorPort.ListenForGetTagHistoryRequest(c, startTS, endTS, histTagDefs)
}
