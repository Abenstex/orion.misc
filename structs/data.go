package structs

import (
	"laniakea/dataStructures"
	"orion.commons/structs"
)

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
	ObjectVersion int
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

type Category struct {
	Info           structs.BaseInfo `json:"info"`
	ReferencedType string           `json:"referencedType"`
}

type CategoryReference struct {
	CategoryId    uint64 `json:"categoryId"`
	ObjectId      uint64 `json:"objectId"`
	ObjectType    string `json:"objectType"`
	ObjectVersion int    `json:"objectVersion"`
}

type ObjectTypeCustomization struct {
	Id                dataStructures.JsonNullInt64  `json:"id"`
	ObjectType        string                        `json:"objectType"`
	FieldName         string                        `json:"fieldName"`
	FielDataType      string                        `json:"fieldDataType"`
	FieldMandatory    bool                          `json:"fieldMandatory"`
	FieldDefaultValue dataStructures.JsonNullString `json:"fieldDefaultValue"`
	CreatedDate       int64                         `json:"createdDate"`
	CreatedBy         dataStructures.JsonNullString `json:"createdBy"`
}
