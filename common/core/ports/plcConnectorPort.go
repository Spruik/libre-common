//the ports package contains all of the interface definitions used as part of the dependency injection proceesin
package ports

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"time"
)

//The PlcConnectorPort defines the functions to be implemented by any connector the PLC (level 2) environment
type PlcConnectorPort interface {

	//Connect is called to invoke the implementation specific activities need to create a client connection
	// to the underlying PLC data source.
	// <connInfo>: a map containing connection information keyed by strings recognized by the implementor
	//             providing, for example, a list of topics for subscription, or tag patterns, etc.
	Connect() error

	//Close is called to close the connection
	Close() error

	//ReadTags takes a list of tag structs and uses the names to fetch their values from the level 2 environment
	// in an implementation-specific way
	//the return list of tag structs contains the values and quality/error information
	ReadTags(inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct

	//WriteTags takes a list of tag structs and uses the name/value information to update the level 2 environment
	// in an implementation-specific way
	//the return list of tag structs contains any error information resulting from the operation
	WriteTags(outTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct

	//ListenForPlcTagChanges is intended to be used in a separate thread where the caller will wait for a message on
	// the provided channel.  The implementor will detect tag changes in an implementation-specific way and
	// provide the tag data to the caller via the channel.  Changes are identified using the changeFiler in an
	// implementation-specific way
	ListenForPlcTagChanges(c chan domain.StdMessageStruct, changeFilter map[string]interface{})

	//GetTagHistory requests all of the changes to the given tags during the specified time range
	GetTagHistory(startTS time.Time, endTS time.Time, inTagDefs []domain.StdMessageStruct) []domain.StdMessageStruct
}
