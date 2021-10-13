package ports

import (
	"errors"
	"fmt"
	"time"

	logging "github.com/Spruik/libre-logging"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

//The LibreHistorianPort interface defines the functions to be provided by the service acting as the history data store in Libre
type LibreHistorianPortDF interface {

	//Connect is called to establish a connection to the data store service
	Connect() error

	//Close is called to close the data store connection
	Close() error

	AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error

	AddEqPropDataPoint(measurement string, eqId string, eqName string, propId string, propName string, propValue interface{}, ts time.Time) error

	QueryRaw(query string) *dataframe.DataFrame

	QueryRecentPointHistory(backTimeToken string, pointName string) *dataframe.DataFrame

	QueryLatestFromPointHistory(pointName string) *dataframe.DataFrame

	//TODO - other query "convenience" methods?
}

type LibreHistorianPortHandler struct {
	logging.LoggingEnabler

	libreHistorianPortDF LibreHistorianPortDF
}

func NewLibreHistorianPortHandler(loggingHook string, libreHistorianPortDF LibreHistorianPortDF) LibreHistorianPortHandler {
	result := LibreHistorianPortHandler{
		libreHistorianPortDF: libreHistorianPortDF,
	}

	if loggingHook == "" {
		loggingHook = "LibreHistorainPortHandler"
	}
	result.SetLoggerConfigHook(loggingHook)

	return result
}

func (h *LibreHistorianPortHandler) Connect() error {
	return h.libreHistorianPortDF.Connect()
}

func (h *LibreHistorianPortHandler) Close() error {
	return h.libreHistorianPortDF.Close()
}

func (h *LibreHistorianPortHandler) AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	return h.libreHistorianPortDF.AddDataPointRaw(measurement, tags, fields, ts)
}

func (h *LibreHistorianPortHandler) AddEqPropDataPoint(measurement string, eqId string, eqName string, propId string, propName string, propValue interface{}, ts time.Time) error {
	return h.libreHistorianPortDF.AddEqPropDataPoint(measurement, eqId, eqName, propId, propName, propValue, ts)
}

func (h *LibreHistorianPortHandler) QueryRaw(query string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianPortHandler executed QueryRaw in %s", time.Since(start))
		h.LogDebug(msg)
	}()
	results := h.libreHistorianPortDF.QueryRaw(query)

	if err := validateDataframe(results); err != nil {
		h.LogError(err)
		results.Err = err
	}

	return results
}

func (h *LibreHistorianPortHandler) QueryRecentPointHistory(backTimeToken string, pointName string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianPortHandler executed QueryRecentPointHistory in %s", time.Since(start))
		h.LogDebug(msg)
	}()
	results := h.libreHistorianPortDF.QueryRecentPointHistory(backTimeToken, pointName)

	if err := validateDataframe(results); err != nil {
		h.LogError(err)
		results.Err = err
	}

	return results
}

func (h *LibreHistorianPortHandler) QueryLatestFromPointHistory(pointName string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianPortHandler executed QueryLatestFromPointHistory in %s", time.Since(start))
		h.LogDebug(msg)
	}()

	results := h.libreHistorianPortDF.QueryRaw(pointName)

	if err := validateDataframe(results); err != nil {
		h.LogError(err)
		results.Err = err
	}

	return results
}

func validateDataframe(df *dataframe.DataFrame) (err error) {
	names := df.Names()
	types := df.Types()

	if len(names) != 5 {
		return errors.New("improper dataframe column count. Expected 5.")
	}

	for i, n := range names {

		switch n {
		case "Tag", "Svalue":
			if types[i] != series.String {
				return errors.New("improper column type. '" + n + "' Expected 'String'")
			}
		case "Dvalue":
			if !(types[i] == series.Int || types[i] == series.Float) {
				return errors.New("improper column type. '" + n + "' Expected 'Int' or 'Float'")
			}
		case "Timestamp", "Quality":
			if types[i] != series.Int {
				return errors.New("improper column type. '" + n + "' Expected 'Int'")
			}
		default:
			return errors.New("improper dataframe column name. '" + n + "'.")
		}
	}

	return nil
}
