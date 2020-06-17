package structs

import (
	"encoding/json"
	"fmt"
	"laniakea/micro"
)

type SaveStatesRequest struct {
	Header        micro.RequestHeader `json:"header"`
	UpdatedStates []State             `json:"updatedStates"`
	OriginalState []State             `json:"originalState"`
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

func (request GetStatesRequest) ToString() (string, error) {
	byteWurst, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error in request ToString" + err.Error())
	}

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
