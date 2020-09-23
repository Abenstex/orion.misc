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
	Header           micro.EventHeader `json:"eventHeader"`
	AttributeChanges []AttributeChange `json:"attributeChange"`
}

func (event AttributeValueChangedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeValueChangedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeValueDeletedEvent struct {
	Header      micro.EventHeader `json:"eventHeader"`
	AttributeId uint64            `json:"attributeId"`
	ObjectId    uint64            `json:"objectId"`
}

func (event AttributeValueDeletedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeValueDeletedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type SavedHierarchiesEvent struct {
	Header      micro.EventHeader `json:"eventHeader"`
	ObjectType  string            `json:"objectType"`
	Hierarchies []Hierarchy       `json:"hierarchies"`
}

func (event SavedHierarchiesEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event SavedHierarchiesEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type ParameterSavedEvent struct {
	Header     micro.EventHeader `json:"eventHeader"`
	ObjectType string            `json:"objectType"`
	Parameters []Parameter       `json:"parameters"`
}

func (event ParameterSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event ParameterSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type CategorySavedEvent struct {
	Header     micro.EventHeader `json:"eventHeader"`
	ObjectType string            `json:"objectType"`
	Categories []Category        `json:"categories"`
}

func (event CategorySavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event CategorySavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type CategoryReferencesSavedEvent struct {
	Header     micro.EventHeader `json:"eventHeader"`
	Categories []uint64          `json:"categories"`
}

func (event CategoryReferencesSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event CategoryReferencesSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}
