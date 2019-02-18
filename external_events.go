package gomarsys

import (
	"encoding/json"
	"fmt"
)

type ExternalEvents struct {
	client ClientInterface
}

type EventData struct {
	ExternalID string            `json:"external_id"`
	Data       map[string]string `json:"data"`
}

type TriggerBatchEvent struct {
	KeyID    int         `json:"key_id"`
	Contacts []EventData `json:"contacts"`
}

type TriggerEvent struct {
	EventData
	KeyID int `json:"key_id"`
}

func NewExternalEvents(client ClientInterface) *ExternalEvents {
	return &ExternalEvents{
		client: client,
	}
}

func (e *ExternalEvents) TriggerEvent(eventId int, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	r := &Request{
		Path:   fmt.Sprintf("/v2/event/%d/trigger", eventId),
		Method: requestPost,
		Body:   data,
	}

	if _, err := e.client.Send(r); err != nil {
		return err
	}

	return nil
}
