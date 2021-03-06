package gomarsys

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestUsers_UpdateUser(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/contact")
		assert.Equal(t, req.Method, RequestMethod(requestPut))
		var v interface{}
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.(map[string]interface{})["key_id"], fmt.Sprintf("%d", EMail))
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", FirstName)], "Test")
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", LastName)], "Test")
		assert.Equal(t, v.(map[string]interface{})["contacts"].([]interface{})[0].(map[string]interface{})[fmt.Sprintf("%d", EMail)], "test@test.ru")
	}).Return([]byte{}, nil)

	user := NewUsers(client)
	user.UpdateUser(User{
		Data: map[int]string{
			FirstName: "Test",
			LastName:  "Test",
			EMail:     "test@test.ru",
		},
	}, EMail)
}

func TestUsers_GetUserInfo(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/contact/getdata")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
		var v interface{}
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.(map[string]interface{})["keyId"], fmt.Sprintf("%d", EMail))
		assert.Equal(t, v.(map[string]interface{})["keyValues"].([]interface{})[0].(string), "test@test.ru")
		assert.Equal(t, v.(map[string]interface{})["fields"].([]interface{})[0].(string), fmt.Sprintf("%d", EMail))
		assert.Equal(t, v.(map[string]interface{})["fields"].([]interface{})[1].(string), fmt.Sprintf("%d", Gender))
	}).Return([]byte(`{"replyCode":0,"replyText":"OK","data":{"errors":[],"result":[{"3":"test@test.ru","5":"2","id":"111111","uid":"fd90tidfpd"}]}}`), nil)

	user := NewUsers(client)
	userData, err := user.GetUserInfo(EMail, "test@test.ru", []int{EMail, Gender})
	require.NoError(t, err)

	assert.Equal(t, userData.Data[Gender], "2")
	assert.Equal(t, userData.Data[EMail], "test@test.ru")
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
		BaseExportRequest: BaseExportRequest{
			DistributionMethod: DistributionMethodLocal,
			ContactFields: []int{
				EMail,
				OptIn,
			},
			Delimiter: ",",
		},
		Origin:   OriginAll,
		OriginID: "0",
		TimeRange: []string{
			time.Now().Format("2006-01-02 15:04"),
			time.Now().Format("2006-01-02 15:04"),
		},
	}

	user := NewUsers(client)
	changes, err := user.GetChanges(request)
	require.NoError(t, err)
	assert.Equal(t, changes.ReplyCode, 0)
	assert.Equal(t, changes.ReplyText, "ok")
	assert.Equal(t, changes.Data.ID, 123)
}

func TestUsers_GetContacts(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/email/getcontacts")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
		var v ContactRequest
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.DistributionMethod, DistributionMethodLocal)
		assert.Equal(t, v.ContactListID, 123)
		assert.Equal(t, v.ContactFields, []int{EMail, OptIn})
	}).Return([]byte(`{"replyCode": 0,"replyText" :"ok","data":{"id": 123}}`), nil)

	request := ContactRequest{
		BaseExportRequest: BaseExportRequest{
			DistributionMethod: DistributionMethodLocal,
			ContactFields: []int{
				EMail,
				OptIn,
			},
			Delimiter: ",",
		},
		ContactListID: 123,
	}

	user := NewUsers(client)
	changes, err := user.GetContacts(request)
	require.NoError(t, err)
	assert.Equal(t, changes.ReplyCode, 0)
	assert.Equal(t, changes.ReplyText, "ok")
	assert.Equal(t, changes.Data.ID, 123)
}

func TestUsers_GetSegments(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/export/filter")
		assert.Equal(t, req.Method, RequestMethod(requestPost))
		var v SegmentRequest
		err := json.NewDecoder(strings.NewReader(string(req.Body))).Decode(&v)
		require.NoError(t, err)

		assert.Equal(t, v.DistributionMethod, DistributionMethodLocal)
		assert.Equal(t, v.Filter, 123)
		assert.Equal(t, v.ContactFields, []int{EMail, OptIn})
	}).Return([]byte(`{"replyCode": 0,"replyText" :"ok","data":{"id": 123}}`), nil)

	request := SegmentRequest{
		BaseExportRequest: BaseExportRequest{
			DistributionMethod: DistributionMethodLocal,
			ContactFields: []int{
				EMail,
				OptIn,
			},
			Delimiter: ",",
		},
		Filter: 123,
	}

	user := NewUsers(client)
	changes, err := user.GetSegment(request)
	require.NoError(t, err)
	assert.Equal(t, changes.ReplyCode, 0)
	assert.Equal(t, changes.ReplyText, "ok")
	assert.Equal(t, changes.Data.ID, 123)
}