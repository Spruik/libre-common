package services_test

import (
	"testing"
	"time"

	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/services"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

type LibreHistorianDFServiceTestCase struct {
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

var LibreHistorianDFServiceTestCases = []LibreHistorianDFServiceTestCase{
	{
		Name:    "Passing Dataframe - Float Dvalue",
		IsError: false,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
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
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
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
		Name:    "Failing Dataframe - Extra Column",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]int{123456}, series.Int, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Tag",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Timestamp",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float64{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Svalue",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]int{123456}, series.Int, "Dvalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Dvalue",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]string{"test"}, series.String, "Svalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Missing Quality",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{123456}, series.Int, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
			series.New([]string{"Extra Column"}, series.String, "An extra Column of data"),
		),
	},
	{
		Name:    "Failing Dataframe - Tag Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]int{1}, series.Int, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Timestamp Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]string{"1634159622"}, series.String, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Quality Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]float64{123.456}, series.Float, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - DValue Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]bool{false}, series.Bool, "Dvalue"),
			series.New([]string{"test"}, series.String, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - SValue Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"Example Tag"}, series.String, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]int{1}, series.Int, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]float32{123.456}, series.Float, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Multiple Wrong Datatype",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]int{1}, series.Int, "Tag"),
			series.New([]int{1634159622}, series.Int, "Timestamp"),
			series.New([]string{"test"}, series.String, "Quality"),
			series.New([]float32{123.456}, series.Float, "Dvalue"),
			series.New([]float32{123.456}, series.Float, "Svalue"),
		),
	},
	{
		Name:    "Failing Dataframe - Multiple Same",
		IsError: true,
		QueryRaw: QueryRaw{
			Query: "test",
		},
		QueryRecentPointHistory: QueryRecentPointHistory{
			BackTimeToken: "test",
			PointName:     "test",
		},
		QueryLatestFromPointHistory: QueryLatestFromPointHistory{
			PointName: "test",
		},
		ExpectedResult: dataframe.New(
			series.New([]string{"test"}, series.String, "Tag"),
			series.New([]string{"test"}, series.String, "Tag"),
			series.New([]string{"test"}, series.String, "Tag"),
			series.New([]string{"test"}, series.String, "Tag"),
			series.New([]string{"test"}, series.String, "Tag"),
		),
	},
}

func TestLibreHistorianDFService(t *testing.T) {

	for _, tc := range LibreHistorianDFServiceTestCases {
		result := &tc.ExpectedResult
		if result == nil {
			t.Errorf("%s: result not defined for test case", tc.Name)
		}
		historinPort := services.NewLibreHistorianDFService("test", MockLibreHistorainDF{
			Result: result,
		})

		if err := historinPort.Connect(); err != nil {
			t.Errorf("%s failed, expected Connect() no error but got %s", tc.Name, err)
		}

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

		if err := historinPort.Close(); err != nil {
			t.Errorf("%s failed, expected Close() no error but got %s", tc.Name, err)
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

func (t MockLibreHistorainDF) AddEqPropDataPoint(point ports.AddEqPropDataPointParams) error {
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
