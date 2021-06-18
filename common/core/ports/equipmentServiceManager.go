package ports

import (
	"sync"
)

////////////////////////////////////////////////////////////////////
// constants for Service Requests
//
//SVCRQST_SHUTDOWN - request a thread shutdown
var SVCRQST_SHUTDOWN = "SHUTDOWN"

//SVCRQST_SHUTDOWN_ACK - acknowledge a thread shutdown request
var SVCRQST_SHUTDOWN_ACK = "SHUTDOWNACK"

//SVCRQST_TAGDATA - request for equipment to handle a tag change
var SVCRQST_TAGDATA = "TAGDATA"

//SVCRQST_TAGDATA - acknowledge that equipment has handled a tag change
var SVCRQST_TAGDATA_ACK = "TAGDATAACK"

////////////////////////////////////////////////////////////////////

//The EquipmentServiceManagerIF interface defines the manager which takes care of the equipment processing threads
type EquipmentServiceManagerIF interface {

	//Initializes the service with a config map and TODO - a set of interfaces for processing tag changes
	Initialize() error
	//Start is called to tell the service to start up the equipment processing environment
	//  the WaitGroup passed in should be used to participate in the caller's thread management
	Start(wg *sync.WaitGroup) error
	//Shutdown is called to tear down and end the processing
	Shutdown() error
}
