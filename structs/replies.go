package structs

import (
	"encoding/json"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/orion.commons/structs"
)

type GetStatesReply struct {
	Header micro.ReplyHeader `json:"header"`
	States []structs.State   `json:"states"`
}

func (reply GetStatesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetStatesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetStatesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetStatesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetAttributeDefinitionsReply struct {
	Header               micro.ReplyHeader             `json:"header"`
	AttributeDefinitions []structs.AttributeDefinition `json:"data"`
}

func (reply GetAttributeDefinitionsReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetAttributeDefinitionsReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetAttributeDefinitionsReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetAttributeDefinitionsReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetAttributeValuesReply struct {
	Header     micro.ReplyHeader   `json:"header"`
	Attributes []structs.Attribute `json:"data"`
}

func (reply GetAttributeValuesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetAttributeValuesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetAttributeValuesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetAttributeValuesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetHierarchiesReply struct {
	Header      micro.ReplyHeader `json:"header"`
	Hierarchies []Hierarchy       `json:"data"`
}

func (reply GetHierarchiesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetHierarchiesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetHierarchiesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetHierarchiesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type EvaluateAttributeReply struct {
	Header micro.ReplyHeader `json:"header"`
	Value  *string           `json:"value"`
}

func (reply EvaluateAttributeReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply EvaluateAttributeReply) Successful() bool {
	return reply.Header.Success
}

func (reply EvaluateAttributeReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply EvaluateAttributeReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetParametersReply struct {
	Header     micro.ReplyHeader `json:"header"`
	Parameters []Parameter       `json:"data"`
}

func (reply GetParametersReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetParametersReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetParametersReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetParametersReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetCategoriesReply struct {
	Header     micro.ReplyHeader `json:"header"`
	Categories []Category        `json:"data"`
}

func (reply GetCategoriesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetCategoriesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetCategoriesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetCategoriesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetObjectsPerCategoriesReply struct {
	Header  micro.ReplyHeader  `json:"header"`
	Objects []structs.BaseInfo `json:"data"`
}

func (reply GetObjectsPerCategoriesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetObjectsPerCategoriesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetObjectsPerCategoriesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetObjectsPerCategoriesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetObjectTypeCustomizationsReply struct {
	Header                   micro.ReplyHeader         `json:"header"`
	ObjectTypeCustomizations []ObjectTypeCustomization `json:"data"`
}

func (reply GetObjectTypeCustomizationsReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetObjectTypeCustomizationsReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetObjectTypeCustomizationsReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetObjectTypeCustomizationsReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetStateTransitionRulesReply struct {
	Header               micro.ReplyHeader     `json:"header"`
	StateTransitionRules []StateTransitionRule `json:"data"`
}

func (reply GetStateTransitionRulesReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetStateTransitionRulesReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetStateTransitionRulesReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetStateTransitionRulesReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}
