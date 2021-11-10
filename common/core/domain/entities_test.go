package domain

import "testing"

func Test_ContainsEquipment(t *testing.T) {

	var equipments = []Equipment{
		{
			Id: "1",
		},
		{
			Id: "2",
		},
		{
			Id: "",
		},
		{
			Id: "3",
		},
		{
			Id: "3",
		},
	}

	if !ContainsEquipment(equipments, Equipment{Id: "1"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 1")
	}

	if !ContainsEquipment(equipments, Equipment{Id: "2"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 2")
	}

	if !ContainsEquipment(equipments, Equipment{Id: "3"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 3")
	}

	if !ContainsEquipment(equipments, Equipment{}) {
		t.Errorf("Expect equipment to contain equipment with out an Id")
	}

	if ContainsEquipment(equipments, Equipment{Id: "0"}) {
		t.Errorf("Expect equipment to NOT to contain equipment with id: 0")
	}

}

func Test_DeduplicateEquipment(t *testing.T) {
	var equipments = []Equipment{
		{
			Id: "1",
		},
		{
			Id: "2",
		},
		{
			Id: "",
		},
		{
			Id: "3",
		},
		{
			Id: "3",
		},
	}

	dedupEquipments := DeduplicateEquipment(equipments)

	if !ContainsEquipment(dedupEquipments, Equipment{Id: "1"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 1")
	}

	if !ContainsEquipment(dedupEquipments, Equipment{Id: "2"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 2")
	}

	if !ContainsEquipment(dedupEquipments, Equipment{Id: "3"}) {
		t.Errorf("Expect equipment to contain equipment with Id: 3")
	}

	if !ContainsEquipment(dedupEquipments, Equipment{}) {
		t.Errorf("Expect equipment to contain equipment with out an Id")
	}

	if len(dedupEquipments) != 4 {
		t.Errorf("Expect deduplicated equipments to contain 4 elements")
	}

}
