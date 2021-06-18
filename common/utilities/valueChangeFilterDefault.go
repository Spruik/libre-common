package utilities

import "github.com/Spruik/libre-common/common/core/domain"

type valueChangeFilterDefault struct {
}

func NewValueChangeFilterDefault() *valueChangeFilterDefault {
	return &valueChangeFilterDefault{}
}

func (s *valueChangeFilterDefault) Initialize() error {
	return nil //no cofig for default
}

func (s *valueChangeFilterDefault) PassValueThrough(tagChange domain.StdMessageStruct) (bool, error) {
	_ = tagChange
	return true, nil //default passes everything through
}
