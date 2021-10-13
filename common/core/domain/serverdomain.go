package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-gota/gota/dataframe"
)

const (
	ADMIN_CMD_START    string = "START"
	ADMIN_CMD_STOP     string = "STOP"
	ADMIN_CMD_SHUTDOWN string = "SHUTDOWN"
	ADMIN_CMD_READY    string = "READY"
)

// TagQuality references to OPC-UA Standard tag quality
type TagQuality int

const (
	// Bad - Non-specific
	Bad TagQuality = 0

	// Uncertain - Non-specific
	Uncertain TagQuality = 84

	// Good - Non-specific
	Good TagQuality = 192
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

// StdMessageStruct used for publishing tag values throughout the libre ecosystem
type StdMessageStruct struct {
	OwningAsset       string                 `json:"OwningAsset"`
	OwningAssetId     string                 `json:"OwningAssetId"`
	ItemName          string                 `json:"ItemName"`
	ItemNameExt       map[string]string      `json:"ItemNameExt"`
	ItemId            string                 `json:"ItemId"`
	ItemValue         interface{}            `json:"ItemValue"`
	ItemOldValue      interface{}            `json:"ItemOldValue"`
	ItemDataType      string                 `json:"ItemDataType"`
	TagQuality        int                    `json:"TagQuality"`
	Err               *string                `json:"Err"`
	ChangedTimestamp  time.Time              `json:"ChangedTimestamp"`
	PreviousTimestamp time.Time              `json:"PreviousTimestamp"`
	Category          string                 `json:"Category"`
	Topic             string                 `json:"Topic"`
	ReplyTopic        string                 `json:"ReplyTopic,omitempty"`
	History           *[]dataframe.DataFrame `json:"History,omitempty"`
}

func ConvertTypes(messageStruct StdMessageStruct) StdMessageStruct {
	switch messageStruct.ItemDataType {
	case "FLOAT":
		messageStruct.ItemValue, _ = strconv.ParseFloat(fmt.Sprintf("%v", messageStruct.ItemValue), 64)
		messageStruct.ItemOldValue, _ = strconv.ParseFloat(fmt.Sprintf("%v", messageStruct.ItemOldValue), 64)
	case "FLOAT64":
		messageStruct.ItemValue, _ = strconv.ParseFloat(fmt.Sprintf("%v", messageStruct.ItemValue), 64)
		messageStruct.ItemOldValue, _ = strconv.ParseFloat(fmt.Sprintf("%v", messageStruct.ItemOldValue), 64)
	case "INT32":
		messageStruct.ItemValue, _ = strconv.Atoi(fmt.Sprintf("%v", messageStruct.ItemValue))
		messageStruct.ItemOldValue, _ = strconv.Atoi(fmt.Sprintf("%v", messageStruct.ItemOldValue))
	case "INT":
		messageStruct.ItemValue, _ = strconv.Atoi(fmt.Sprintf("%v", messageStruct.ItemValue))
		messageStruct.ItemOldValue, _ = strconv.Atoi(fmt.Sprintf("%v", messageStruct.ItemOldValue))
	case "BOOL":
		messageStruct.ItemValue, _ = strconv.ParseBool(fmt.Sprintf("%v", messageStruct.ItemValue))
		messageStruct.ItemOldValue, _ = strconv.ParseBool(fmt.Sprintf("%v", messageStruct.ItemOldValue))
	}
	return messageStruct
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
