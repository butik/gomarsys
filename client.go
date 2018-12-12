package gomarsys

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type auth struct {
	User   string
	Secret string
}

const maxLengthWSSE = 32

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const host = "https://api.emarsys.net/api/"

const (
	requestPost = iota + 1
	requestGet
	requestPut
)

type RequestMethod int

type Client struct {
	auth auth
	host string
}

type Request struct {
	Path   string
	Method RequestMethod
	Body   []byte
}

type ClientInterface interface {
	Send(r *Request) ([]byte, error)
}

func NewClient(user string, secret string) ClientInterface {
	return &Client{
		auth: auth{
			User:   user,
			Secret: secret,
		},
		host: host,
	}
}

func (c *Client) Send(r *Request) ([]byte, error) {
	method := ""

	switch r.Method {
	case requestPost:
		method = "POST"
	case requestGet:
		method = "GET"
	case requestPut:
		method = "PUT"
	default:
		return nil, fmt.Errorf("unknown method: %d", r.Method)
	}

	serverUrl, _ := url.Parse(c.host)
	serverUrl.Path += strings.TrimLeft(r.Path, "/")

	req, err := http.NewRequest(method, serverUrl.String(), bytes.NewBuffer(r.Body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-WSSE", c.getWSSEHeader())
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response: %s", responseBody)
	}

	return responseBody, nil
}

func (c *Client) getWSSEHeader() string {
	b := make([]byte, maxLengthWSSE)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	nonce := string(b)

	var timestamp = time.Now().Format(time.RFC3339)
	text := nonce + timestamp + c.auth.Secret
	h := sha1.New()
	h.Write([]byte(text))
	s := hex.EncodeToString(h.Sum(nil))
	passwordDigest := base64.StdEncoding.EncodeToString([]byte(s))
	path.Join()

	wsse := []string{
		fmt.Sprintf("Username=\"%s\"", c.auth.User),
		fmt.Sprintf("PasswordDigest=\"%s\"", passwordDigest),
		fmt.Sprintf("Nonce=\"%s\"", nonce),
		fmt.Sprintf("Created=\"%s\"", timestamp),
	}

	return fmt.Sprintf(" UsernameToken %s", strings.Join(wsse, ","))
}
