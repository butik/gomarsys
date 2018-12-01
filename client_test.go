package gomarsys

import (
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Send(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/some/test/path")
		assert.Equal(t, req.Method, "POST")
		assert.Matches(t, req.Header.Get("X-WSSE"), `^UsernameToken Username="test",PasswordDigest="[\w=]+",Nonce="[\w=]+",Created="[0-9\-+:T]+"$`)
		assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
	}))
	defer server.Close()

	client := &Client{
		auth: auth{
			User:   "test",
			Secret: "test",
		},

		host: server.URL,
	}

	r := &Request{
		Path:   "/some/test/path",
		Method: requestPost,
		Body:   []byte("OK"),
	}

	_, err := client.Send(r)
	require.NoError(t, err)
}
