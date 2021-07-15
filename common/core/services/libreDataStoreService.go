package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
)

type libreDataStoreService struct {
	port ports.LibreDataStorePort
}

func NewLibreDataStoreService(storeIF ports.LibreDataStorePort) *libreDataStoreService {
	return &libreDataStoreService{
		port: storeIF,
	}
}

var libreDataStoreServiceInstance *libreDataStoreService = nil

func SetLibreDataStoreServiceInstance(inst *libreDataStoreService) {
	libreDataStoreServiceInstance = inst
}
func GetLibreDataStoreServiceInstance() *libreDataStoreService {
	return libreDataStoreServiceInstance
}

func (s *libreDataStoreService) Connect() error {
	return s.port.Connect()
}
func (s *libreDataStoreService) Close() error {
	return s.port.Close()
}
func (s *libreDataStoreService) BeginTransaction(forUpdate bool, name string) ports.LibreDataStoreTransactionPort {
	return s.port.BeginTransaction(forUpdate, name)
}
func (s *libreDataStoreService) GetSubscription(q interface{}, vars map[string]interface{}) ports.LibreDataStoreSubscriptionPort {
	return s.port.GetSubscription(q, vars)
}
