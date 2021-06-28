package drivers

import (
	"context"
	"fmt"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

type libreHistorianInfluxdb struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler

	org                   string
	bucket                string
	client                influxdb2.Client
	writeAPI              api.WriteAPIBlocking
	queryAPI              api.QueryAPI
	eqPropMeasurementName string
}

func NewLibreHistorianInfluxdb(configHook string) *libreHistorianInfluxdb {
	s := libreHistorianInfluxdb{}
	s.SetConfigCategory(configHook)
	hook, err := s.GetConfigItemWithDefault("loggerHook", "Influxdb")
	if err == nil {
		s.SetLoggerConfigHook(hook)
	}
	s.SetLoggerConfigHook(hook)
	return &s
}

func (s *libreHistorianInfluxdb) Connect() error {
	var err error
	var url string
	var authToken string
	url, err = s.GetConfigItem("serverURL")
	if err == nil {
		authToken, err = s.GetConfigItem("authToken")
		if err == nil {
			s.org, err = s.GetConfigItem("org")
			if err == nil {
				s.bucket, err = s.GetConfigItem("bucket")
				if err != nil {
					panic(fmt.Sprintf("Failed to get the 'bucket' entry from configuration [%s]", err))
				}
			} else {
				panic(fmt.Sprintf("Failed to get the 'org' entry from configuration [%s]", err))
			}
		} else {
			panic(fmt.Sprintf("Failed to get the 'authToken' entry from configuration [%s]", err))
		}
	} else {
		panic(fmt.Sprintf("Failed to get the 'serverURL' entry from configuration [%s]", err))
	}
	if err == nil {
		s.client = influxdb2.NewClient(url, authToken)
		s.writeAPI = s.client.WriteAPIBlocking(s.org, s.bucket)
		s.queryAPI = s.client.QueryAPI(s.org)
	} else {
		panic(fmt.Sprintf("Failed in configuration of the InfluxDB client: %s", err))
	}
	return err
}

func (s *libreHistorianInfluxdb) Close() error {
	if s.client != nil {
		s.client.Close()
	}
	return nil //influx close returns no value
}

func (s *libreHistorianInfluxdb) AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	// Create point using full params constructor
	p := influxdb2.NewPoint(measurement,
		tags,
		fields,
		ts)
	// write point immediately
	return s.writeAPI.WritePoint(context.Background(), p)
}

func (s *libreHistorianInfluxdb) AddEqPropDataPoint(measurement string, eqId string, eqName string, propId string, propName string, propValue interface{}, ts time.Time) error {
	// Create point using fluent style
	p := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("equipmentId", eqId).
		AddTag("propId", propId).
		AddField(propName, propValue).
		SetTime(ts)
	s.LogDebugf("using eqId=%s, eqName=%s, propId=%s, propName=%s, propValue=%+v, ts=%+v", eqId, eqName, propId, propName, propValue, ts)
	s.LogDebugf("built a new point for influxdb storage of a prop value:  %+v", p)
	return s.writeAPI.WritePoint(context.Background(), p)
}

func (s *libreHistorianInfluxdb) QueryRaw(query string) (*api.QueryTableResult, error) {
	return s.queryAPI.Query(context.Background(), query)
}

func (s *libreHistorianInfluxdb) QueryRecentPointHistory(backTimeToken string, pointName string) (*api.QueryTableResult, error) {
	query := fmt.Sprintf(`from(bucket:"%s")|> range(start: %s) |> filter(fn: (r) => r._measurement == "%s")`, s.bucket, backTimeToken, pointName)
	return s.queryAPI.Query(context.Background(), query)
}
