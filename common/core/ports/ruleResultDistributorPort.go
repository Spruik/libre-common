package ports

type RuleResultDistributorIF interface {
	DistributeRuleResult(mgdEq *ManagedEquipmentPort, ruleResults map[string]interface{}) error
}
