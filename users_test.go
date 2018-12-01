package gomarsys

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
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
