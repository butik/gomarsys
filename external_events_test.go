package gomarsys

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExternalEvents_TriggerEvent(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/event/1/trigger")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
	}).Return([]byte{}, nil)

	externalEvent := NewExternalEvents(client)
	externalEvent.TriggerEvent(1, TriggerEvent{
		EventData: EventData{
			ExternalID: "some@client.ru",
			Data: map[string]string{
				"some_var": "some_val",
			},
		},
		KeyID: 1,
	})
}

func TestExternalEvents_TriggerBatchEvent(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/event/1/trigger")
		assert.Equal(t, req.Method, RequestMethod(requestPost))

		var data map[string]interface{}
		err := json.Unmarshal(req.Body, &data)
		require.NoError(t, err)
		assert.Len(t, data["contacts"].([]interface{}), 1)
	}).Return([]byte{}, nil)

	externalEvent := NewExternalEvents(client)
	externalEvent.TriggerEvent(1, TriggerBatchEvent{
		Contacts: []EventData{
			{
				ExternalID: "some@client.ru",
				Data: map[string]string{
					"some_var": "some_val",
				},
			},
		},
		KeyID: 1,
	})
}
