package services

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"time"
)

type plcConnectorService struct {
	plcConnectorPort ports.PlcConnectorPort
}

func NewPlcConnectorService(port ports.PlcConnectorPort) *plcConnectorService {
	return &plcConnectorService{
		plcConnectorPort: port,
	}
}

var plcConnectorServiceInstance *plcConnectorService = nil

func SetPlcConnectorServiceInstance(inst *plcConnectorService) {
	plcConnectorServiceInstance = inst
}
func GetPlcConnectorServiceInstance() *plcConnectorService {
	return plcConnectorServiceInstance
}

func (s *plcConnectorService) Connect() error {
	return s.plcConnectorPort.Connect()
}
func (s *plcConnectorService) Close() error {
	return s.plcConnectorPort.Close()
}
func (s *plcConnectorService) ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.plcConnectorPort.ReadTags(inTagDefs)
}
func (s *plcConnectorService) WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.plcConnectorPort.WriteTags(outTagDefs)
}
func (s *plcConnectorService) ListenForPlcTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{}) {
	s.plcConnectorPort.ListenForPlcTagChanges(c, changeFilter)
}
func (s *plcConnectorService) Unsubscribe( equipmentId *string,topicList []string) error{
	return s.plcConnectorPort.Unsubscribe(equipmentId,topicList)
}
func (s *plcConnectorService) GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct {
	return s.plcConnectorPort.GetTagHistory(startTS, endTS, inTagDefs)
}

//////////////////////////////////////////////////////////////////
type plcConnectorFactoryService struct {
	port ports.PlcConnectorFactoryPort
}

var plcConnectorFactoryServiceInstance *plcConnectorFactoryService = nil

func SetPlcConnectorFactoryServiceInstance(inst *plcConnectorFactoryService) {
	plcConnectorFactoryServiceInstance = inst
}
func GetPlcConnectorFactoryServiceInstance() *plcConnectorFactoryService {
	return plcConnectorFactoryServiceInstance
}

func NewPlcConnectorFactoryService(port ports.PlcConnectorFactoryPort) *plcConnectorFactoryService {
	var ret = plcConnectorFactoryService{}
	ret.port = port
	return &ret
}

func (s *plcConnectorFactoryService) CreatePlcConnectorInstance(key string, configHook string) ports.PlcConnectorPort {
	return s.port.CreatePlcConnectorInstance(key, configHook)
}
