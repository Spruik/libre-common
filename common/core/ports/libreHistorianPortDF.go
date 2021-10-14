package ports

import (
	"time"

	"github.com/go-gota/gota/dataframe"
)

// LibreHistorianPortDF interface defines the functions to be provided by the service acting as the history data store in Libre
type LibreHistorianPortDF interface {

	// Connect is called to establish a connection to the data store service
	Connect() error

	// Close is called to close the data store connection
	Close() error

	AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error

	AddEqPropDataPoint(measurement string, eqId string, eqName string, propId string, propName string, propValue interface{}, ts time.Time) error

	QueryRaw(query string) *dataframe.DataFrame

	QueryRecentPointHistory(backTimeToken string, pointName string) *dataframe.DataFrame

	QueryLatestFromPointHistory(pointName string) *dataframe.DataFrame
}
