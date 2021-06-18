package ports

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"sync"
)

//The EquipmentServiceManagerRunnerIF interface defines the handler function called by the EquipmentServiceManager
//  for each "live" equipment (in a separate thread)
type EquipmentServiceManagerRunnerIF interface {
	Prepare(eqId string)
	Run(rqstChannel chan domain.EquipmentServiceRequest, wg *sync.WaitGroup)
}
