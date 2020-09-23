package structs

import (
	"encoding/json"
	"laniakea/micro"
	"orion.commons/couchdb"
	"orion.commons/structs"
)

type GetStatesReply struct {
	Header micro.ReplyHeader `json:"header"`
	States []State           `json:"states"`
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
	AttributeDefinitions []structs.AttributeDefinition `json:"attributeDefinitions"`
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
	Attributes []structs.Attribute `json:"attributes"`
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

type GetAttributeValueChangeHistoryReply struct {
	Header            micro.ReplyHeader                         `json:"header"`
	ChangedAttributes []couchdb.HistoricizedAttributeDataChange `json:"changedAttributes"`
}

func (reply GetAttributeValueChangeHistoryReply) MarshalJSON() (string, error) {
	bytes, err := json.Marshal(reply)

	return string(bytes), err
}

func (reply GetAttributeValueChangeHistoryReply) Successful() bool {
	return reply.Header.Success
}

func (reply GetAttributeValueChangeHistoryReply) Error() string {
	if reply.Header.ErrorMessage != nil {
		return *reply.Header.ErrorMessage
	}

	return ""
}

func (reply GetAttributeValueChangeHistoryReply) GetHeader() *micro.ReplyHeader {
	return &reply.Header
}

type GetHierarchiesReply struct {
	Header      micro.ReplyHeader `json:"header"`
	Hierarchies []Hierarchy       `json:"hierarchies"`
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
	Parameters []Parameter       `json:"parameters"`
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
	Categories []Category        `json:"categories"`
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
	Objects []structs.BaseInfo `json:"objects"`
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
