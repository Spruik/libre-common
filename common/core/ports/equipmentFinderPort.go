package ports

import "github.com/Spruik/libre-common/common/core/domain"

type EquipmentFinderChangeNotice struct {
	ChangeType string
	Equipment  []domain.Equipment
}

//The EquipmentFinderPort interface defines the functions to support the finding of equipment to support
type EquipmentFinderPort interface {

	//FindEquipment locates equipment that should be active
	FindEquipment() ([]domain.Equipment, error)

	//SubscribeToChanges - Be informed about changes in the equipemnt matching selection in the configuration
	SubscribeToChanges(notificationChannel chan EquipmentFinderChangeNotice)

	//UnsubscribeToChanges will unsubscribe (called with the same channel as the SubscribeToChanges call)
	UnsubscribeToChanges(notificationChannel chan EquipmentFinderChangeNotice)
}
