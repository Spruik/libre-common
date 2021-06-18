package utilities

import (
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/Spruik/libre-common/common/core/queries"
)

type plcEqNameResolverAlias struct {
	dataStore ports.LibreDataStorePort
	system    string
}

func NewPlcEqNameResolverAlias(dataStore ports.LibreDataStorePort, aliasingSystem string) *plcEqNameResolverAlias {
	return &plcEqNameResolverAlias{
		dataStore: dataStore,
		system:    aliasingSystem,
	}
}

func (s *plcEqNameResolverAlias) ResolvePlcEqName(plcName string) (string, error) {
	//use data store alias info to translate name
	return queries.GetEquipmentNameForSystemAlias(s.dataStore.BeginTransaction(false, "eqaliasplc"), s.system, plcName)
}

func (s *plcEqNameResolverAlias) ResolveStdEqName(stdName string) (string, error) {
	return queries.GetAliasEquipmentNameForSystem(s.dataStore.BeginTransaction(false, "eqaliasstd"), s.system, stdName)
}
