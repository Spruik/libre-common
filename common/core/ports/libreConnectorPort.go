//the ports package contains all of the interface definitions used as part of the dependency injection proceesin
package ports

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"time"
)

//The LibreConnectorPort defines the functions to be implemented by any connector to the PLC (level 2) environment
type LibreConnectorPort interface {

	//Connect is called to invoke the implementation specific activities need to create a client connection
	// to the Libre Central Service.
	Connect() error

	//Close is called to close the connection
	Close() error

	//SendTagChange sends the data for a changed tag to the Libre Central Service
	//OUTGOING: called from EdgeAgent to send to LibreCentral
	SendStdMessage(msg domain.StdMessageStruct) error

	//ListenForReadTagsRequest takes a list of tag structs and uses Edge Agent facilities to get the tag values
	//the return list of tag structs contains the values and quality/error information
	ListenForReadTagsRequest(c chan []domain.StdMessageStruct, readTagDefs []domain.StdMessageStruct)

	//ListenForWriteTagsRequest takes a list of tag structs and uses Edge Agent facilities to set the tag values
	// in an implementation-specific way
	//the return list of tag structs contains any error information resulting from the operation
	ListenForWriteTagsRequest(c chan []domain.StdMessageStruct, writeTagDefs []domain.StdMessageStruct)

	//ListenForGetTagHistoryRequest requests all of the changes to the given tags during the specified time range
	//INCOMING:  message by LibreCentral to invoke activity ih EdgeAgent
	ListenForGetTagHistoryRequest(c chan []domain.StdMessageStruct, startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct)
}
