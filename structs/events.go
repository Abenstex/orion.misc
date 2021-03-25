package structs

import (
	"encoding/json"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/orion.commons/structs"
)

type SavedStatesEvent struct {
	Header     micro.EventHeader `json:"event_header"`
	ObjectType string            `json:"object_type"`
	States     []structs.State   `json:"states"`
}

func (event SavedStatesEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event SavedStatesEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type SavedStateTransitionRulesEvent struct {
	Header               micro.EventHeader     `json:"event_header"`
	ObjectType           string                `json:"object_type"`
	StateTransitionRules []StateTransitionRule `json:"state_transition_rules"`
}

func (event SavedStateTransitionRulesEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event SavedStateTransitionRulesEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeDefinitionSavedEvent struct {
	Header               micro.EventHeader             `json:"event_header"`
	ObjectType           string                        `json:"object_type"`
	AttributeDefinitions []structs.AttributeDefinition `json:"attribute_definitions"`
}

func (event AttributeDefinitionSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeDefinitionSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeValueChangedEvent struct {
	Header           micro.EventHeader `json:"event_header"`
	AttributeChanges []AttributeChange `json:"attribute_change"`
}

func (event AttributeValueChangedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeValueChangedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type AttributeValueDeletedEvent struct {
	Header      micro.EventHeader `json:"event_header"`
	AttributeId uint64            `json:"attribute_id"`
	ObjectId    uint64            `json:"object_id"`
}

func (event AttributeValueDeletedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event AttributeValueDeletedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type SavedHierarchiesEvent struct {
	Header      micro.EventHeader `json:"event_header"`
	ObjectType  string            `json:"object_type"`
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
	Header     micro.EventHeader `json:"event_header"`
	ObjectType string            `json:"object_type"`
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
	Header     micro.EventHeader `json:"event_header"`
	ObjectType string            `json:"object_type"`
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
	Header     micro.EventHeader `json:"event_header"`
	Categories []uint64          `json:"categories"`
}

func (event CategoryReferencesSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event CategoryReferencesSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}

type ObjectTypeCustomizationsSavedEvent struct {
	Header                   micro.EventHeader `json:"event_header"`
	ObjectTypeCustomizations []string          `json:"object_type_customizations"`
	ObjectType               string            `json:"object_type"`
}

func (event ObjectTypeCustomizationsSavedEvent) ToJsonString() (string, error) {
	byteWurst, err := json.Marshal(event)

	return string(byteWurst), err
}

func (event ObjectTypeCustomizationsSavedEvent) GetHeader() micro.EventHeader {
	return event.Header
}
