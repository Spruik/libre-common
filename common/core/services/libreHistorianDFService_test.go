package services_test

import (
	"testing"
	"time"

	"github.com/Spruik/libre-common/common/core/services"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

type LibreHistorianPortHandlerTestCase struct {
	Name     string
	IsError  bool
	QueryRaw struct {
		Query string
	}
	QueryRecentPointHistory struct {
		BackTimeToken string
		PointName     string
	}
	QueryLatestFromPointHistory struct {
		PointName string
	}
	ExpectedResult dataframe.DataFrame
}

var LibreHistorianPortHandlerTestCases = []LibreHistorianPortHandlerTestCase{
	{
		Name:    "Passing Dataframe - Float Dvalue",
		IsError: false,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Passing Dataframe - Int Dvalue",
		IsError: false,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]int{123456}, series.Int, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Tag",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Timestamp",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float64{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Svalue",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]int{123456}, series.Int, "Dvalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Dvalue",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Quality",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "Test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "Test",
			PointName:     "Test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "Test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{123456}, series.Int, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
}

func TestLibreHistorianPortHandler(t *testing.T) {

	for _, tc := range LibreHistorianPortHandlerTestCases {
		result := &tc.ExpectedResult
		if result == nil {
			t.Errorf("%s: result not defined for test case", tc.Name)
		}
		historinPort := services.NewLibreHistorianPortHandler("test", MockLibreHistorainDF{
			Result: result,
		})

		if tc.QueryRaw.Query != "" {
			result := historinPort.QueryRaw(tc.QueryRaw.Query)

			if tc.IsError && result.Err == nil {
				t.Errorf("%s failed, expected an error but got nil", tc.Name)
			} else if !tc.IsError && result.Err != nil {
				t.Errorf("%s failed, expected no error but got %s", tc.Name, result.Err)
			}
		}

		if tc.QueryLatestFromPointHistory.PointName != "" {
			result := historinPort.QueryLatestFromPointHistory(tc.QueryLatestFromPointHistory.PointName)
			if tc.IsError && result.Err == nil {
				t.Errorf("%s failed, expected an error but got nil", tc.Name)
			} else if !tc.IsError && result.Err != nil {
				t.Errorf("%s failed, expected no error but got %s", tc.Name, result.Err)
			}
		}

		if tc.QueryRecentPointHistory.PointName != "" {
			result := historinPort.QueryRecentPointHistory(tc.QueryRecentPointHistory.BackTimeToken, tc.QueryRecentPointHistory.PointName)
			if tc.IsError && result.Err == nil {
				t.Errorf("%s failed, expected an error but got nil", tc.Name)
			} else if !tc.IsError && result.Err != nil {
				t.Errorf("%s failed, expected no error but got %s", tc.Name, result.Err)
			}
		}
	}
}

type QueryRaw struct {
	Query string
}

type QueryRecentPointHistory struct {
	BackTimeToken string
	PointName     string
}

type QueryLatestFromPointHistory struct {
	PointName string
}

// MockLibreHistorainDF is a mock instance of the LibreHistorianDF
type MockLibreHistorainDF struct {
	Result *dataframe.DataFrame
}

func (t MockLibreHistorainDF) Connect() error {
	return nil
}

func (t MockLibreHistorainDF) Close() error {
	return nil
}

func (t MockLibreHistorainDF) AddDataPointRaw(measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	return nil
}

func (t MockLibreHistorainDF) AddEqPropDataPoint(measurement string, eqId string, eqName string, propId string, propName string, propValue interface{}, ts time.Time) error {
	return nil
}

func (t MockLibreHistorainDF) QueryRaw(query string) *dataframe.DataFrame {
	return t.Result
}

func (t MockLibreHistorainDF) QueryRecentPointHistory(backTimeToken string, pointName string) *dataframe.DataFrame {
	return t.Result
}

func (t MockLibreHistorainDF) QueryLatestFromPointHistory(pointName string) *dataframe.DataFrame {
	return t.Result
}
