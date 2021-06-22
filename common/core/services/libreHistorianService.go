package services

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

type libreHistorianService struct {
	port ports.LibreHistorianPort
}

func NewLibreHistorianService(port ports.LibreHistorianPort) *libreHistorianService {
	return &libreHistorianService{
		port: port,
	}
}

var libreHistorianServiceInstance *libreHistorianService = nil

func SetLibreHistorianServiceInstance(inst *libreHistorianService) {
	libreHistorianServiceInstance = inst
}
func GetLibreHistorianServiceInstance() *libreHistorianService {
	return libreHistorianServiceInstance
}

func (s *libreHistorianService) Connect() error {
	return s.port.Connect()
}
func (s *libreHistorianService) Close() error {
	return s.port.Close()
}
func (s *libreHistorianService) AddDataPointRaw(pointName string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	return s.port.AddDataPointRaw(pointName, tags, fields, ts)
}
func (s *libreHistorianService) AddEqPropDataPoint(eqId string, eqName string, propId string, propName string, ts time.Time) error {
	return s.port.AddEqPropDataPoint(eqId, eqName, propId, propName, ts)
}
func (s *libreHistorianService) QueryRaw(query string) (*api.QueryTableResult, error) {
	return s.port.QueryRaw(query)
}
func (s *libreHistorianService) QueryRecentPointHistory(backTimeToken string, pointName string) (*api.QueryTableResult, error) {
	return s.port.QueryRecentPointHistory(backTimeToken, pointName)
}
