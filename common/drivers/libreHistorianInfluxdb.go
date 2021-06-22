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
	hook, err := s.GetConfigItemWithDefault("loggingHook", "Influxdb")
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
				if err == nil {
					s.eqPropMeasurementName, err = s.GetConfigItem("eqPropMeasurementName")
				}
			}
		}
	}
	if err == nil {
		if err == nil {
			s.client = influxdb2.NewClient(url, authToken)
			s.writeAPI = s.client.WriteAPIBlocking(s.org, s.bucket)
			s.queryAPI = s.client.QueryAPI(s.org)
		}
	} else {
		panic(fmt.Sprintf("Failed in configuration of the InfluxDB client: %s", err))
	}
	return err
}

func (s *libreHistorianInfluxdb) Close() error {
	return nil
}

func (s *libreHistorianInfluxdb) AddDataPointRaw(pointName string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	// Create point using full params constructor
	p := influxdb2.NewPoint(pointName,
		tags,
		fields,
		ts)
	// write point immediately
	return s.writeAPI.WritePoint(context.Background(), p)
}

func (s *libreHistorianInfluxdb) AddEqPropDataPoint(eqId string, eqName string, propId string, propName string, ts time.Time) error {
	// Create point using fluent style
	p := influxdb2.NewPointWithMeasurement(s.eqPropMeasurementName).
		AddTag("equipmentId", eqId).
		AddField("propId", 23.2).
		SetTime(ts)
	return s.writeAPI.WritePoint(context.Background(), p)
}

func (s *libreHistorianInfluxdb) QueryRaw(query string) (*api.QueryTableResult, error) {
	return s.queryAPI.Query(context.Background(), query)
}

func (s *libreHistorianInfluxdb) QueryRecentPointHistory(backTimeToken string, pointName string) (*api.QueryTableResult, error) {
	query := fmt.Sprintf(`from(bucket:"%s")|> range(start: %s) |> filter(fn: (r) => r._measurement == "%s")`, s.bucket, backTimeToken, pointName)
	return s.queryAPI.Query(context.Background(), query)
}
