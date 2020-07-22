package structs

import (
	"encoding/json"
	"laniakea/micro"
	"orion.commons/structs"
)

type SaveStatesRequest struct {
	Header        micro.RequestHeader `json:"header"`
	UpdatedStates []State             `json:"updatedStates"`
	OriginalState []State             `json:"originalState"`
}

func (request *SaveStatesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveStatesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *SaveStatesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request SaveStatesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type GetStatesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"whereClause"`
}

func (request *GetStatesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetStatesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetStatesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request GetStatesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveStateTransitionRulesRequest struct {
	Header                      micro.RequestHeader   `json:"header"`
	UpdatedStateTransitionRules []StateTransitionRule `json:"updatedStateTransitionRules"`
}

func (request *SaveStateTransitionRulesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveStateTransitionRulesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *SaveStateTransitionRulesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request SaveStateTransitionRulesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type DefineAttributeRequest struct {
	Header                      micro.RequestHeader           `json:"header"`
	UpdatedAttributeDefinitions []structs.AttributeDefinition `json:"updatedAttributeDefinitions"`
	OriginalAttributeDefintions []structs.AttributeDefinition `json:"originalAttributeDefinition"`
}

func (request *DefineAttributeRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request DefineAttributeRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *DefineAttributeRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request DefineAttributeRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type GetAttributeDefinitionsRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"whereClause"`
}

func (request *GetAttributeDefinitionsRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetAttributeDefinitionsRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetAttributeDefinitionsRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request GetAttributeDefinitionsRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SetAttributeValueRequest struct {
	Header      micro.RequestHeader `json:"header"`
	AttributeId uint64              `json:"attributeId"`
	Value       string              `json:"value"`
	ObjectType  string              `json:"objectType"`
	ObjectId    uint64              `json:"objectId"`
}

func (request *SetAttributeValueRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SetAttributeValueRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *SetAttributeValueRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		error := reply.Error()
		header.ExecutionError = &error
	}
	request.Header = header

	return request
}

func (request SetAttributeValueRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}
