package gomarsys

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestUsers_Create(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/contact")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
		var v interface{}
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.(map[string]interface{})["key_id"], fmt.Sprintf("%d", EMail))
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", FirstName)], "Test")
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", LastName)], "Test")
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", EMail)], "test@test.ru")
	}).Return([]byte{}, nil)

	user := NewUsers(client)
	user.Create(User{
		Data: map[int]string{
			FirstName: "Test",
			LastName:  "Test",
			EMail:     "test@test.ru",
		},
	}, EMail)
}

func TestUsers_GetChanges(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/contact/getchanges")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
		var v ChangesRequest
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.Origin, OriginAll)
		assert.Equal(t, v.DistributionMethod, DistributionMethodLocal)
		assert.Equal(t, v.OriginID, "0")
		assert.Equal(t, v.ContactFields, []int{EMail, OptIn})
	}).Return([]byte(`{"replyCode": 0,"replyText" :"ok","data":{"id": 123}}`), nil)

	request := ChangesRequest{
		DistributionMethod: DistributionMethodLocal,
		Origin:             OriginAll,
		OriginID:           "0",
		TimeRange: []string{
			time.Now().Format("2006-01-02 15:04"),
			time.Now().Format("2006-01-02 15:04"),
		},
		ContactFields: []int{
			EMail,
			OptIn,
		},
		Delimiter: ",",
	}

	user := NewUsers(client)
	changes, err := user.GetChanges(request)
	require.NoError(t, err)
	assert.Equal(t, changes.ReplyCode, 0)
	assert.Equal(t, changes.ReplyText, "ok")
	assert.Equal(t, changes.Data.ID, 123)
}
