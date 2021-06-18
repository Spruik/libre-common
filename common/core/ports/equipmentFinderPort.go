package ports

import "github.com/Spruik/libre-common/common/core/domain"

//The EquipmentFinderPort interface defines the functions to support the finding of equipment to support
type EquipmentFinderPort interface {

	//FindEquipment locates equipment that should be active
	FindEquipment() ([]domain.Equipment, error)
}
