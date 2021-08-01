package ports

//The EquipmentCachePort interface defines the functions to support the caching of equipment references
type EquipmentCachePort interface {

	//RefreshCache builds/rebuilds the internal cache of equipment references based on the given configuration map
	RefreshCache()
	//GetEquipmentCacheList returns the entire cache as a map keyed by the equipment name
	GetEquipmentCacheList() *map[string]*ManagedEquipmentPort
	//GetCachedEquipmentItem returns the managed equipment structure keyed by the given equipment name
	GetCachedEquipmentItem(equipName string) *ManagedEquipmentPort
	GetCachedEquipmentItemById(equipId string) *ManagedEquipmentPort
	SetEquipmentChangeNoticeFunction(handlingFxn func(notice EquipmentCacheChangeNotice))
	StartMonitoring()
	StopMonitoring()
}

type EquipmentCacheChangeNotice struct {
	ChangeType string //TODO - should probably be an enum
	EqId       string
}
