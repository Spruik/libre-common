package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The ValueChangeFilterPort interface describes the functions to be provided by any value filter
type ValueChangeFilterPort interface {
	Initialize() error

	PassValueThrough(tagChange domain.StdMessageStruct) (bool, error)
}

//////////////////////////////////

type ValueChangeFilterFactoryPort interface {
	CreateFilterInstance(key string, mgdEq *ManagedEquipmentPort) ValueChangeFilterPort
}
