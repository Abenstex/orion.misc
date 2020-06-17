package structs

import (
	"encoding/json"
	"laniakea/micro"
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