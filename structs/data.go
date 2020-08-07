package structs

import "orion.commons/structs"

type State struct {
	Info            structs.BaseInfo `json:"info"`
	ReferencedType  string           `json:"referencedType"`
	ObjectAvailable bool             `json:"objectAvailable"`
	Substate        bool             `json:"substate"`
	DefaultState    bool             `json:"defaultState"`
	Substates       []int64          `json:"subStates"`
}

type StateTransitionRule struct {
	Info      structs.BaseInfo `json:"info"`
	FromState int64            `json:"fromState"`
	ToStates  []int64          `json:"toStates"`
}

type AttributeChange struct {
	AttributeId   uint64
	ObjectType    string
	ObjectId      uint64
	OriginalValue string
	NewValue      string
}

type Hierarchy struct {
	Info    structs.BaseInfo `json:"info"`
	Entries []HierarchyEntry `json:"entries"`
}

func NewHierarchy() Hierarchy {
	hierarchy := Hierarchy{Entries: make([]HierarchyEntry, 0)}

	return hierarchy
}

type HierarchyEntry struct {
	Index      int    `json:"index"`
	ObjectType string `json:"objectType"`
}

type Parameter struct {
	Info  structs.BaseInfo `json:"info"`
	Value string           `json:"value"`
}
