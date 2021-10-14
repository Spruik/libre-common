package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/Spruik/libre-common/common/core/ports"
	libreLogger "github.com/Spruik/libre-logging"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

// LibreHistorianDFService wraps a LibreHistorianPortDF with validation of resulting queries and performance metrics.
type LibreHistorianDFService struct {
	libreLogger.LoggingEnabler
	libreHistorianPortDF ports.LibreHistorianPortDF
	islibreLogger        bool
}

// NewLibreHistorianDFService creates a new LibreHistorianDFService with a logger hook and LibreHistorianPortDF
func NewLibreHistorianDFService(loggingHook string, libreHistorianPortDF ports.LibreHistorianPortDF) (result LibreHistorianDFService) {
	result = LibreHistorianDFService{
		islibreLogger:        true,
		libreHistorianPortDF: libreHistorianPortDF,
	}

	if loggingHook == "" {
		loggingHook = "LibreHistorainPortHandlesr"
	}

	defer func() {
		_ = recover()
		result = LibreHistorianDFService{
			islibreLogger:        false,
			libreHistorianPortDF: libreHistorianPortDF,
		}
	}()

	result.SetLoggerConfigHook(loggingHook)

	return result
}

// Connect calls Connect on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) Connect() error {
	return h.libreHistorianPortDF.Connect()
}

// Close calls Close on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) Close() error {
	return h.libreHistorianPortDF.Close()
}

// AddDataPointRaw calls AddDataPointRaw on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	return h.libreHistorianPortDF.AddDataPointRaw(measurement, tags, fields, ts)
}

// AddEqPropDataPoint calls AddEqPropDataPoint on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) AddEqPropDataPoint(measurement string, eqID string, eqName string, propID string, propName string, propValue interface{}, ts time.Time) error {
	return h.libreHistorianPortDF.AddEqPropDataPoint(measurement, eqID, eqName, propID, propName, propValue, ts)
}

// QueryRaw calls QueryRaw on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) QueryRaw(query string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianDFService executed QueryRaw in %s", time.Since(start))
		if h.IsLibreLogger() {
			h.LogDebug(msg)
		} else {
			fmt.Println(msg)
		}
	}()
	results := h.libreHistorianPortDF.QueryRaw(query)

	if err := validateDataframe(results); err != nil {
		if h.IsLibreLogger() {
			h.LogError(err)
		} else {
			fmt.Println(err)
		}
		results.Err = err
	}

	return results
}

// QueryRecentPointHistory calls QueryRecentPointHistory on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) QueryRecentPointHistory(backTimeToken string, pointName string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianDFService executed QueryRecentPointHistory in %s", time.Since(start))
		if h.IsLibreLogger() {
			h.LogDebug(msg)
		} else {
			fmt.Println(msg)
		}
	}()
	results := h.libreHistorianPortDF.QueryRecentPointHistory(backTimeToken, pointName)

	if err := validateDataframe(results); err != nil {
		if h.IsLibreLogger() {
			h.LogError(err)
		} else {
			fmt.Println(err)
		}
		results.Err = err
	}

	return results
}

// QueryLatestFromPointHistory calls QueryLatestFromPointHistory on the LibreHistorianPortDF instance
func (h *LibreHistorianDFService) QueryLatestFromPointHistory(pointName string) *dataframe.DataFrame {
	start := time.Now()
	defer func() {
		msg := fmt.Sprintf("LibreHistorianDFService executed QueryLatestFromPointHistory in %s", time.Since(start))
		if h.IsLibreLogger() {
			h.LogDebug(msg)
		} else {
			fmt.Println(msg)
		}

	}()

	results := h.libreHistorianPortDF.QueryRaw(pointName)

	if err := validateDataframe(results); err != nil {
		if h.IsLibreLogger() {
			h.LogError(err)
		} else {
			fmt.Println(err)
		}
		results.Err = err
	}

	return results
}

// IsLibreLogger returns true when using the LibreLogger
func (h *LibreHistorianDFService) IsLibreLogger() bool {
	return h.islibreLogger
}

// Validate Result contains columns "Tag", "Timestamp", "Svalue", "Dvalue", "Quality"
func validateDataframe(df *dataframe.DataFrame) error {
	names := df.Names()
	types := df.Types()

	if len(names) != 5 {
		return errors.New("improper dataframe column count. Expected 5")
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
