package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const (
	ADMIN_CMD_START    string = "START"
	ADMIN_CMD_STOP     string = "STOP"
	ADMIN_CMD_SHUTDOWN string = "SHUTDOWN"
	ADMIN_CMD_READY    string = "READY"
)

// OpcUaQuality refers to OPC-UA Standard tag quality
type OpcUaQuality int

const (
	// Bad - Non-specific
	Bad OpcUaQuality = 0

	// Uncertain - Non-specific
	Uncertain OpcUaQuality = 84

	// Good - Non-specific
	Good OpcUaQuality = 192
)

type AdminCommand struct {
	Cmd  string
	Args map[string]interface{}
}

type DateTime string
type StdMessage struct {
	Topic   string
	Payload *json.RawMessage
}


// TimeseriesValue is a representation of a value in a point in time, uses OPC-UA Standard for Tag Quality
type TimeseriesValue struct {
	Value     interface{}
	Timestamp time.Time
	Quality   OpcUaQuality
}

// StdMessageStruct used for publishing tag values throughout the libre ecosystem
type StdMessageStruct struct {
	OwningAsset       string             `json:"OwningAsset"`
	OwningAssetId     string             `json:"OwningAssetId"`
	ItemName          string             `json:"ItemName"`
	ItemNameExt       map[string]string  `json:"ItemNameExt"`
	ItemId            string             `json:"ItemId"`
	ItemValue         string             `json:"ItemValue"`
	ItemOldValue      string             `json:"ItemOldValue"`
	ItemDataType      string             `json:"ItemDataType"`
	TagQuality        int                `json:"TagQuality"`
	Err               *string            `json:"Err"`
	ChangedTimestamp  time.Time          `json:"ChangedTimestamp"`
	PreviousTimestamp time.Time          `json:"PreviousTimestamp"`
	Category          string             `json:"Category"`
	Topic             string             `json:"Topic"`
	ReplyTopic        string             `json:"ReplyTopic,omitempty"`
	History           *[]TimeseriesValue `json:"History,omitempty"`
}

type EquipmentPropertyDescriptor struct {
	Name             string
	DataType         string
	Value            interface{}
	ClassPropertyId  string
	EquipmentClassId string
	LastUpdate       time.Time
}

type EquipmentEventDescriptor struct {
	Name   string
	Time   time.Time
	Params map[string]interface{}
}

type EquipmentServiceRequest struct {
	ServiceType string
	Time        time.Time
	Message     string
	TagInfo     StdMessageStruct
}

const (
	SVCRQST_SHUTDOWN     = "SHUTDOWN"
	SVCRQST_SHUTDOWN_ACK = "SHUTDOWNACK"
	SVCRQST_TAGDATA      = "TAGDATA"
	SVCRQST_TAGDATA_ACK  = "TAGDATAACK"
)

///////////////////////////////////////////////////

func ConvertPropertyValueStringToTypedValue(propType string, rawVal interface{}) (interface{}, error) {
	var val interface{}
	var err error = nil
	switch rawVal.(type) {
	case string:
		var strVal string = fmt.Sprintf("%s", rawVal)
		switch propType {
		case "STRING":
			val = rawVal
		case "BOOL":
			if strVal == "" {
				val = false
			} else {
				val, err = strconv.ParseBool(strVal)
			}
		case "INT":
			if strVal == "" {
				val = int(0)
			} else {
				val, err = strconv.ParseInt(strVal, 10, 16)
			}
		case "INT32":
			if strVal == "" {
				val = int32(0)
			} else {
				val, err = strconv.ParseInt(strVal, 10, 32)
			}
		case "FLOAT64":
			if strVal == "" {
				val = float64(0.0)
			} else {
				val, err = strconv.ParseFloat(strVal, 64)
			}
		case "FLOAT":
			if strVal == "" {
				val = float32(0.0)
			} else {
				val, err = strconv.ParseFloat(strVal, 32)
			}
		}
	default:
		val = rawVal
	}
	return val, err
}
