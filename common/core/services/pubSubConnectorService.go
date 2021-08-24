package services

import (
	"encoding/json"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
)

type pubSubConnectorService struct {
	port ports.PubSubConnectorPort
}

func NewPubSubConnectorService(port ports.PubSubConnectorPort) *pubSubConnectorService {
	return &pubSubConnectorService{
		port: port,
	}
}

var pubSubConnectorServiceInstance *pubSubConnectorService = nil

func SetPubSubConnectorServiceInstance(inst *pubSubConnectorService) {
	pubSubConnectorServiceInstance = inst
}
func GetPubSubConnectorServiceInstance() *pubSubConnectorService {
	return pubSubConnectorServiceInstance
}

func (s *pubSubConnectorService) Connect() error {
	return s.port.Connect()
}
func (s *pubSubConnectorService) Close() error {
	return s.port.Close()
}
func (s *pubSubConnectorService) Publish(topic string, payload *json.RawMessage, qos byte, retain bool) error{
	return s.port.Publish(topic,payload,qos,retain)
}
func (s *pubSubConnectorService) Subscribe(c chan *domain.StdMessage, topicMap map[string]string)  {
	s.port.Subscribe(c,topicMap)
}
