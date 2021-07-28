package utilities

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
)

type plcTagResolverAlias struct {
	dataStore ports.LibreDataStorePort
	system    string
}

func NewPlcTagResolverAlias(dataStore ports.LibreDataStorePort, aliasingSystem string) *plcTagResolverAlias {
	return &plcTagResolverAlias{
		dataStore: dataStore,
		system:    aliasingSystem,
	}
}

func (s *plcTagResolverAlias) ResolvePlcTagName(plcName string, eqName string) (string, error) {
	//use data store alias info to translate name
	name, err := queries.GetPropertyNameForSystemAlias(s.dataStore.BeginTransaction(false, "tagaliasplc"), s.system, plcName, eqName)
	if err == nil && name == "" {
		//no translation - return original name
		name = plcName
	}
	return name, err
}

func (s *plcTagResolverAlias) ResolveStdTagName(stdName domain.StdMessageStruct) (string, error) {
	ret, err := queries.GetAliasPropertyNameForSystem(s.dataStore.BeginTransaction(false, "tagaliasstd"), s.system, stdName.ItemName, stdName.OwningAsset)
	if err == nil && ret == "" {
		//no translation - return original name
		ret = stdName.ItemName
	}
	return ret, err
}
