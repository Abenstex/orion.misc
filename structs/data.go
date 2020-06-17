package structs

import "orion.commons/structs"

type State struct {
	Info            structs.BaseInfo `json:"info"`
	ReferencedType  string           `json:"referencedType"`
	ObjectAvailable bool             `json:"objectAvailable"`
	Substate        bool             `json:"substate"`
	DefaultState    bool             `json:"defaultState"`
}
