package queries

import (
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	"github.com/hasura/go-graphql-client"
)

func GetActiveEquipmentByLevel(txn ports.LibreDataStoreTransactionPort, level string) ([]domain.Equipment, error) {
	var q struct {
		QueryEquipment []domain.Equipment `graphql:"queryEquipment (filter:{isActive: true, and: {equipmentLevel: {in: [$level]}}}) "`
	}
	type EquipmentElementLevel string
	variables := map[string]interface{}{
		"level": EquipmentElementLevel(level),
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.QueryEquipment, err
}

func GetActiveEquipmentByLevelList(txn ports.LibreDataStoreTransactionPort, levels []domain.EquipmentElementLevel) ([]domain.Equipment, error) {
	var q struct {
		QueryEquipment []domain.Equipment `graphql:"queryEquipment (filter:{isActive: true, and: {equipmentLevel: {in: $levels}}}) "`
	}
	variables := map[string]interface{}{
		"levels": levels,
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.QueryEquipment, err
}

func GetActiveEquipmentByLevelListWithIncExc(txn ports.LibreDataStoreTransactionPort, levels []domain.EquipmentElementLevel, includeIds []string, excludeIds []string) ([]domain.Equipment, error) {
	var q struct {
		QueryEquipment []domain.Equipment `graphql:"queryEquipment (filter:{isActive: true, and: {equipmentLevel: {in: $levels}, or: {id:$includeIds}, and: {not:{id:$excludeIds}}}}) "`
	}
	variables := map[string]interface{}{
		"levels":     levels,
		"includeIds": includeIds,
		"excludeIds": excludeIds,
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.QueryEquipment, err
}

func GetEquipmentByName(txn ports.LibreDataStoreTransactionPort, eqName string) (domain.Equipment, error) {
	var q struct {
		QueryEquipment []domain.Equipment `graphql:"queryEquipment (filter:{name:{eq:$eqName}}) "`
	}
	variables := map[string]interface{}{
		"eqName": graphql.String(eqName),
	}
	err := txn.ExecuteQuery(&q, variables)
	if len(q.QueryEquipment) > 0 {
		return q.QueryEquipment[0], err
	}
	return domain.Equipment{}, err
}

func GetEquipmentById(txn ports.LibreDataStoreTransactionPort, eqId string) (domain.Equipment, error) {
	var q struct {
		GetEquipment domain.Equipment `graphql:"getEquipment(id:$eqId) " json:"getEquipment"`
	}
	var variables = map[string]interface{}{
		"eqId": graphql.ID(eqId),
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.GetEquipment, err
}
func GetEquipmentClassPropertiesAndParentById(txn ports.LibreDataStoreTransactionPort, eqcId string) (domain.EquipmentClassPropertiesAndParent, error) {
	var q struct {
		GetEquipmentClass domain.EquipmentClassPropertiesAndParent `graphql:"getEquipmentClass(id:$eqcId) " json:"getEquipmentClass"`
	}
	var variables = map[string]interface{}{
		"eqcId": graphql.ID(eqcId),
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.GetEquipmentClass, err
}

func GetEquipmentClassById(txn ports.LibreDataStoreTransactionPort, eqcId string) (domain.EquipmentClass, error) {
	var q struct {
		GetEquipmentClass domain.EquipmentClass `graphql:"getEquipmentClass(id:$eqcId) " json:"getEquipmentClass"`
	}
	var variables = map[string]interface{}{
		"eqcId": graphql.ID(eqcId),
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.GetEquipmentClass, err
}

func GetEquipmentNameForSystemAlias(txn ports.LibreDataStoreTransactionPort, system string, plcName string) (string, error) {
	var q struct {
		QueryEquipmentNameAlias []struct {
			Equipment struct {
				Name string
			}
			Alias  string
			System string
		} `graphql:"queryEquipmentNameAlias(filter:{alias: {alloftext: $plcName} system: {alloftext: $system}})"`
	}
	variables := map[string]interface{}{
		"plcName": graphql.String(plcName),
		"system":  graphql.String(system),
	}
	err := txn.ExecuteQuery(&q, variables)
	if len(q.QueryEquipmentNameAlias) > 0 {
		return q.QueryEquipmentNameAlias[0].Equipment.Name, err
	}
	return "", err
}

func GetAliasEquipmentNameForSystem(txn ports.LibreDataStoreTransactionPort, system string, stdName string) (string, error) {
	var q struct {
		QueryEquipmentNameAlias []struct {
			Equipment struct {
				Name string
			} `graphql:"equipment(filter:{name:{eq:$stdName}})"`
			Alias  string
			System string
		} `graphql:"queryEquipmentNameAlias(filter:{system: {alloftext: $system}}) @cascade "`
	}
	variables := map[string]interface{}{
		"stdName": graphql.String(stdName),
		"system":  graphql.String(system),
	}
	err := txn.ExecuteQuery(&q, variables)
	if len(q.QueryEquipmentNameAlias) > 0 {
		return q.QueryEquipmentNameAlias[0].Alias, err
	}
	return "", err
}

func GetPropertyNameForSystemAlias(txn ports.LibreDataStoreTransactionPort, system string, externalName string, eqName string) (string, error) {
	var q struct {
		QueryPropertyNameAlias []struct {
			Property struct {
				Name string
			}
			Equipment struct {
				Name string
			} `graphql:"equipment(filter:{name: {eq: $eqName}})"`
			Alias  string
			System string
		} `graphql:"queryPropertyNameAlias(filter:{alias: {alloftext: $extName} system: {alloftext: $system}})"`
	}
	variables := map[string]interface{}{
		"eqName":  graphql.String(eqName),
		"extName": graphql.String(externalName),
		"system":  graphql.String(system),
	}
	err := txn.ExecuteQuery(&q, variables)
	if len(q.QueryPropertyNameAlias) > 0 {
		return q.QueryPropertyNameAlias[0].Property.Name, err
	}
	return "", err
}

func GetAliasPropertyNamesForSystem(txn ports.LibreDataStoreTransactionPort, system string, eqName string) (map[string]string, error) {
	q := struct {
		QueryPropertyNameAlias []struct {
			Property struct {
				Id   string `json:"id"`
				Name string `json:"name"`
			}
			Equipment struct {
				Name string
			} `graphql:"equipment(filter:{name: {eq: $eqName}})"`
			Alias  string
			System string
		} `graphql:"queryPropertyNameAlias(filter:{system: {alloftext: $system}}) @cascade "`
	}{}
	variables := map[string]interface{}{
		"eqName": graphql.String(eqName),
		"system": graphql.String(system),
	}
	err := txn.ExecuteQuery(&q, variables)
	if err == nil {
		ret := map[string]string{}
		for _, j := range q.QueryPropertyNameAlias {
			ret[j.Property.Name] = j.Alias
		}
		return ret, nil
	}
	return nil, err
}

func GetAliasPropertyNameForSystem(txn ports.LibreDataStoreTransactionPort, system string, internalName string, eqName string) (string, error) {
	eq, err := GetEquipmentByName(txn, eqName)
	if err == nil {
		var propMap map[string]domain.Property
		propMap, err = GetAllPropertiesForEquipment(txn, eq.Id)
		if err == nil {
			targetProp := propMap[internalName]
			q := struct {
				QueryPropertyNameAlias []struct {
					Property struct {
						Id   string `json:"id"`
						Name string `json:"name"`
					} `graphql:"property(filter:{id:[$propId]})"`
					Equipment struct {
						Name string
					} `graphql:"equipment(filter:{name: {eq: $eqName}})"`
					Alias  string
					System string
				} `graphql:"queryPropertyNameAlias(filter:{system: {alloftext: $system}}) @cascade "`
			}{}
			variables := map[string]interface{}{
				"eqName": graphql.String(eqName),
				"propId": graphql.ID(targetProp.Id),
				"system": graphql.String(system),
			}
			err := txn.ExecuteQuery(&q, variables)
			if len(q.QueryPropertyNameAlias) > 0 {
				return q.QueryPropertyNameAlias[0].Alias, err
			}
		}
	}
	return "", err
}

func GetAllPropertiesForEquipment(txn ports.LibreDataStoreTransactionPort, eqId string) (map[string]domain.Property, error) {
	//need to look for properties attached to the Equipment and to it's equipment class (and equipment class parents)
	var fullPropertyList = map[string]domain.Property{}
	//first check equipment
	var eqInst domain.Equipment
	var err error = nil
	eqInst, err = GetEquipmentById(txn, eqId)
	if err == nil {
		for _, p := range eqInst.Properties {
			fullPropertyList[p.Name] = p
		}

		//now work up through the equipment classes
		currEqcId := eqInst.EquipmentClass.Id
		var eqcInst domain.EquipmentClassPropertiesAndParent
		for currEqcId != "" {
			eqcInst, err = GetEquipmentClassPropertiesAndParentById(txn, currEqcId)
			if err == nil {
				for _, p := range eqcInst.Properties {
					fullPropertyList[p.Name] = p
				}
				currEqcId = eqcInst.Parent.Id
			} else {
				currEqcId = ""
			}
		}
	}
	return fullPropertyList, err
}

func GetEventDefinitionById(txn ports.LibreDataStoreTransactionPort, eventDefId string) (domain.EventDefinition, error) {
	var q struct {
		GetEventDefinition domain.EventDefinition `graphql:"getEventDefinition(id:$eventDefId) " json:"getEventDefinition"`
	}
	var variables = map[string]interface{}{
		"eventDefId": graphql.ID(eventDefId),
	}
	err := txn.ExecuteQuery(&q, variables)
	return q.GetEventDefinition, err
}

func GetAllEventDefsForEquipmentAndClass(txn ports.LibreDataStoreTransactionPort, eqId string) ([]domain.EventDefinition, error) {
	//need to look for event defs attached to the EquipmentClass or any parent EquipmentClass
	var fullEventDefList []domain.EventDefinition = nil
	//first check equipment
	var eqInst domain.Equipment
	var err error = nil
	eqInst, err = GetEquipmentById(txn, eqId)
	if err == nil {
		var eqcInst domain.EquipmentClass
		var currEqcParentId = eqInst.EquipmentClass.Id
		for currEqcParentId != "" {
			eqcInst, err = GetEquipmentClassById(txn, currEqcParentId)
			if err == nil {
				for _, e := range eqcInst.EventDefinitions {
					fullEventDefList = append(fullEventDefList, e)
				}
				eqcInst, err = GetEquipmentClassById(txn, currEqcParentId)
				if err == nil {
					currEqcParentId = eqcInst.Parent.Id
				}
			} else {
				break
			}
		}
	}
	return fullEventDefList, err
}

func GetEquipmentElementLevels(txn ports.LibreDataStoreTransactionPort, eventDefId string) ([]string, error) {
	var q struct {
		Levels struct {
			EnumValues []struct {
				Name string `json:"name"`
			} `json:"enumValues"`
		} `graphql:"__type(name:\"EquipmentElementLevel\")" json:"__type"`
	}
	err := txn.ExecuteQuery(&q, nil)
	ret := make([]string, 0, 0)
	for _, j := range q.Levels.EnumValues {
		ret = append(ret, j.Name)
	}
	return ret, err
}
