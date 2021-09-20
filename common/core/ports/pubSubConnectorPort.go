package ports

import (
	"encoding/json"
	"github.com/Spruik/libre-common/common/core/domain"
)

type PubSubConnectorPort interface {
	Connect() error
	Close() error
	Publish(topic string, payload *json.RawMessage, qos byte, retain bool, username *string) error
	Subscribe(c chan *domain.StdMessage, topicMap map[string]string, changeFilter map[string]interface{})
}
