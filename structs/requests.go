package structs

import (
	"encoding/json"
	"github.com/abenstex/laniakea/micro"
	"github.com/abenstex/orion.commons/structs"
)

type SaveStatesRequest struct {
	Header        micro.RequestHeader `json:"header"`
	UpdatedStates []structs.State     `json:"updated_states"`
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
		err := reply.Error()
		header.ExecutionError = &err
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
	WhereClause *string             `json:"where_clause"`
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
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetStatesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveStateTransitionRulesRequest struct {
	Header                      micro.RequestHeader   `json:"header"`
	UpdatedStateTransitionRules []StateTransitionRule `json:"updated_state_transition_rules"`
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
		err := reply.Error()
		header.ExecutionError = &err
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
	UpdatedAttributeDefinitions []structs.AttributeDefinition `json:"updated_attribute_definitions"`
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
		err := reply.Error()
		header.ExecutionError = &err
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
	WhereClause *string             `json:"where_clause"`
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
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetAttributeDefinitionsRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SetAttributeValueRequest struct {
	Header     micro.RequestHeader `json:"header"`
	Attributes []structs.Attribute `json:"attributes"`
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
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request SetAttributeValueRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type DeleteAttributeValueRequest struct {
	Header      micro.RequestHeader `json:"header"`
	AttributeId uint64              `json:"attribute_id"`
	ObjectId    uint64              `json:"object_id"`
}

func (request *DeleteAttributeValueRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request DeleteAttributeValueRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *DeleteAttributeValueRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request DeleteAttributeValueRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetAttributeValuesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	AttributeId *uint64             `json:"attribute_id"`
	ObjectId    *uint64             `json:"object_id"`
}

func (request *GetAttributeValuesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetAttributeValuesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetAttributeValuesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetAttributeValuesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetAttributeValueChangeHistoryRequest struct {
	Header   micro.RequestHeader `json:"header"`
	ObjectId uint64              `json:"object_id"`
}

func (request *GetAttributeValueChangeHistoryRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetAttributeValueChangeHistoryRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetAttributeValueChangeHistoryRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetAttributeValueChangeHistoryRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveHierarchiesRequest struct {
	Header             micro.RequestHeader `json:"header"`
	UpdatedHierarchies []Hierarchy         `json:"updated_hierarchies"`
}

func (request *SaveHierarchiesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveHierarchiesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *SaveHierarchiesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request SaveHierarchiesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type GetHierarchiesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetHierarchiesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetHierarchiesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetHierarchiesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetHierarchiesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type EvaluateAttributeRequest struct {
	Header      micro.RequestHeader `json:"header"`
	ObjectId    int64               `json:"object_id"`
	AttributeId int64               `json:"attribute_id"`
}

func (request *EvaluateAttributeRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request EvaluateAttributeRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *EvaluateAttributeRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request EvaluateAttributeRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveParametersRequest struct {
	Header     micro.RequestHeader `json:"header"`
	Parameters []Parameter         `json:"parameters"`
}

func (request *SaveParametersRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveParametersRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *SaveParametersRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request SaveParametersRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type GetParametersRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetParametersRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetParametersRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetParametersRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetParametersRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetModuleReadmesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetModuleReadmesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetModuleReadmesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetModuleReadmesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetModuleReadmesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveCategoriesRequest struct {
	Header            micro.RequestHeader `json:"header"`
	UpdatedCategories []Category          `json:"updated_categories"`
}

func (request *SaveCategoriesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveCategoriesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

func (request *SaveCategoriesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request SaveCategoriesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

type GetCategoriesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetCategoriesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetCategoriesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetCategoriesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetCategoriesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetObjectsPerCategoriesRequest struct {
	Header     micro.RequestHeader `json:"header"`
	CategoryId *int64              `json:"category_id"`
}

func (request *GetObjectsPerCategoriesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetObjectsPerCategoriesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetObjectsPerCategoriesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetObjectsPerCategoriesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type SaveObjectTypeCustomizationsRequest struct {
	Header                   micro.RequestHeader       `json:"header"`
	ObjectTypeCustomizations []ObjectTypeCustomization `json:"object_type_customizations"`
}

func (request *SaveObjectTypeCustomizationsRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request SaveObjectTypeCustomizationsRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *SaveObjectTypeCustomizationsRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request SaveObjectTypeCustomizationsRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetObjectTypeCustomizationsRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetObjectTypeCustomizationsRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetObjectTypeCustomizationsRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetObjectTypeCustomizationsRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetObjectTypeCustomizationsRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}

type GetStateTransitionRulesRequest struct {
	Header      micro.RequestHeader `json:"header"`
	WhereClause *string             `json:"where_clause"`
}

func (request *GetStateTransitionRulesRequest) UpdateHeader(header *micro.RequestHeader) {
	request.Header = *header
}

func (request GetStateTransitionRulesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)

	return string(byteWurst), err
}

func (request *GetStateTransitionRulesRequest) HandleResult(reply micro.IReply) micro.IRequest {
	header := request.Header
	header.WasExecutedSuccessfully = reply.Successful()
	if len(reply.Error()) > 0 {
		err := reply.Error()
		header.ExecutionError = &err
	}
	request.Header = header

	return request
}

func (request GetStateTransitionRulesRequest) GetHeader() *micro.RequestHeader {
	return &request.Header
}
