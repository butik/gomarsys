package gomarsys

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		KeyID:      1,
		ExternalID: "some@client.ru",
		Data: map[string]string{
			"some_var": "some_val",
		},
	})
}
