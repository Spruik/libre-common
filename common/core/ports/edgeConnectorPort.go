//the ports package contains all of the interface definitions used as part of the dependency injection proceesin
package ports

import (
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
)

//The EdgeConnectorPort defines the functions to be implemented by any connector to the PLC (level 2) environment
type EdgeConnectorPort interface {

	//Connect is called to invoke the implementation specific activities need to create a client connection
	// to the Libre Central Service.
	// <connInfo>: a map containing connection information keyed by strings recognized by the implementor
	Connect(clientId string) error

	//Close is called to close the connection
	Close() error

	//SendTagChange sends the data for a changed tag to the Libre Central Service
	//OUTGOING: called from EdgeAgent to send to LibreCentral
	SendStdMessage(msg domain.StdMessageStruct) error

	//ReadTags takes a list of tag structs and uses the names to fetch their values from the level 2 environment
	// in an implementation-specific way
	//the return list of tag structs contains the values and quality/error information
	ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct

	//WriteTags takes a list of tag structs and uses the name/value information to update the level 2 environment
	// in an implementation-specific way
	//the return list of tag structs contains any error information resulting from the operation
	WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct

	// ListenForEdgeTagChanges is intended to be used in a separate thread where the caller will wait for a message on
	// the provided channel.  The implementor will detect tag changes in an implementation-specific way and
	// provide the tag data to the caller via the channel.  Changes are identified using the changeFilter in an
	// implementation-specific way
	//
	// T.H. - What the heck is changeFilter?
	// I've only seen it used with the keys "EQ" and "Client". If you have a "Client" you can subscribe multiple times
	// to separate channels, oooh nice! Otherwise calling this twice panics the application ¯\_(ツ)_/¯ indicating you can
	// only subscribe to a single channel. Oh trying to do both panics it as well.
	//
	// The "EQ" key, is just the equipment you want to subscribe to for tag changes. E.g. "Site/Area/Line"
	ListenForEdgeTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{})

	// Removes the subscription called from ListenForEdgeTagChanges for a specific client.
	// A client is specified when calling ListenForEdgeTagChanges with a changeFilter key of
	// "Client" specifying the client as a string.
	// This finds the corresponding subscription and channel reference and closes it.
	StopListeningForTagChanges(client string) error

	//GetTagHistory requests all of the changes to the given tags during the specified time range
	GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct
}
