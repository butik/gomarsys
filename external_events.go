package gomarsys

import (
	"encoding/json"
	"fmt"
)

type ExternalEvents struct {
	client ClientInterface
}

type TriggerEvent struct {
	KeyID      int               `json:"key_id"`
	ExternalID string            `json:"external_id"`
	Data       map[string]string `json:"data"`
}

func NewExternalEvents(client ClientInterface) *ExternalEvents {
	return &ExternalEvents{
		client: client,
	}
}

func (e *ExternalEvents) TriggerEvent(eventId int, event TriggerEvent) error {
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
