package structs

import (
	"encoding/json"
	"laniakea/micro"
	"orion.commons/structs"
)

type SavedStatesEvent struct {
	Header     micro.EventHeader `json:"eventHeader"`
	ObjectType string            `json:"objectType"`
	States     []State           `json:"states"`
}

func (event SavedStatesEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event SavedStatesEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type SavedStateTransitionRulesEvent struct {
	Header               micro.EventHeader     `json:"eventHeader"`
	ObjectType           string                `json:"objectType"`
	StateTransitionRules []StateTransitionRule `json:"stateTransitionRules"`
}

func (event SavedStateTransitionRulesEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event SavedStateTransitionRulesEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeDefinitionSavedEvent struct {
	Header               micro.EventHeader             `json:"eventHeader"`
	ObjectType           string                        `json:"objectType"`
	AttributeDefinitions []structs.AttributeDefinition `json:"attributeDefinitions"`
}

func (event AttributeDefinitionSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeDefinitionSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeValueChangedEvent struct {
	Header      micro.EventHeader `json:"eventHeader"`
	ObjectType  string            `json:"objectType"`
	ObjectId    uint64            `json:"objectId"`
	AttributeId uint64            `json:"attributeId"`
	Value       string            `json:"value"`
}

func (event AttributeValueChangedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeValueChangedEvent) GetHeader() micro.EventHeader {
	return event.Header
}
