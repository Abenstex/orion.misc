package structs

import (
	"encoding/json"
	"laniakea/micro"
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
