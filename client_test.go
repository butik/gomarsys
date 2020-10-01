package gomarsys

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), "/some/test/path/")
		assert.Equal(t, req.Method, "POST")
		assert.Regexp(t, `^UsernameToken Username="test",PasswordDigest="[\w=]+",Nonce="[\w=]+",Created="[0-9\-+:T]+"$`, req.Header.Get("X-WSSE"))
		assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
	}))
	defer server.Close()

	client := NewClient("test", "test", WithCustomHost(server.URL + "/"))

	r := &Request{
		Path:   "/some/test/path",
		Method: requestPost,
		Body:   []byte("OK"),
	}

	_, err := client.Send(r)
	require.NoError(t, err)
}
