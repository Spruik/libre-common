package domain

type IdNameTypenameRef struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	TypeName string `json:"__typename" graphql:"__typename"`
}

type IdTypenameRef struct {
	Id       string `json:"id"`
	TypeName string `json:"__typename" graphql:"__typename"`
}

type Property struct {
	Id             string            `json:"id"`
	Name           string            `json:"name"`
	Value          string            `json:"value"`
	DataType       string            `json:"dataType"`
	Equipment      IdNameTypenameRef `json:"equipment"`
	EquipmentClass IdNameTypenameRef `json:"equipmentClass"`
}

type EquipmentElementLevel string

type EquipmentStub struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	EquipmentLevel EquipmentElementLevel `json:"equipmentLevel"`
}

type Equipment struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Properties     []Property            `json:"properties"`
	EquipmentClass IdNameTypenameRef     `json:"equipmentClass"`
	EquipmentLevel EquipmentElementLevel `json:"equipmentLevel"`
	Parent         IdNameTypenameRef     `json:"parent"`
}

type EquipmentClassPropertiesAndParent struct {
	Id         string            `json:"id"`
	Properties []Property        `json:"properties"`
	Parent     IdNameTypenameRef `json:"parent"`
}

type EquipmentClassStub struct {
	Id               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Properties       []Property            `json:"properties"`
	EventDefinitions []IdNameTypenameRef   `json:"eventDefinitions"`
	Parent           IdNameTypenameRef     `json:"parent"`
	EquipmentLevel   EquipmentElementLevel `json:"equipmentLevel"`
}

type EquipmentClass struct {
	Id               string                `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Properties       []Property            `json:"properties"`
	EventDefinitions []EventDefinition     `graphql:"eventDefinitions(filter:{has:name})" json:"eventDefinitions"`
	Parent           IdNameTypenameRef     `json:"parent"`
	EquipmentLevel   EquipmentElementLevel `json:"equipmentLevel"`
}

type MessageClass string
type EventDefinition struct {
	Id                string       `json:"id"`
	Name              string       `json:"name"`
	MessageClass      MessageClass `json:"messageClass"`
	TriggerProperties []Property   `json:"triggerProperties"`
	TriggerExpression string       `json:"triggerExpression"`
	PayloadFields     []struct {
		Name       string `json:"name"`
		Expression string `json:"expression"`
		FieldType  string `json:"fieldType"`
	} `json:"payloadFields"`
	PayloadProperties []Property `json:"payloadProperties"`
}
