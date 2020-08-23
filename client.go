package gomarsys

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
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

const defaultHost = "https://api.emarsys.net/api/"

const mysqlDateFormat = "2006-01-02"

const (
	requestPost = iota + 1
	requestGet
	requestPut
)

type ClientOptions func(o *options)

type RequestMethod int

type ClientError struct {
	err error
}

type options struct {
	client *http.Client
	host   string
}

func WithCustomHost(host string) ClientOptions {
	return func(options *options) {
		options.host = strings.TrimRight(host, "/") + "/"
	}
}

func WithClientTimeout(t time.Duration) ClientOptions {
	return func(options *options) {
		options.client.Timeout = t
	}
}

func (clientError *ClientError) Error() string {
	return clientError.err.Error()
}

func newClientError(err error) error {
	return &ClientError{err}
}

type Client struct {
	auth   auth
	client *http.Client
	host   string
}

type Request struct {
	Path   string
	Method RequestMethod
	Body   []byte
}

type ClientInterface interface {
	Send(r *Request) ([]byte, error)
	SendIO(r *Request) (io.ReadCloser, error)
}

func NewClient(user string, secret string, clientOptions ...ClientOptions) ClientInterface {
	defaultOptions := &options{
		client: http.DefaultClient,
		host:   strings.TrimRight(defaultHost, "/") + "/",
	}

	for _, f := range clientOptions {
		f(defaultOptions)
	}

	return &Client{
		auth: auth{
			User:   user,
			Secret: secret,
		},
		client: defaultOptions.client,
		host:   defaultOptions.host,
	}
}

func (c *Client) SendIO(r *Request) (io.ReadCloser, error) {
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

	serverUrl, err := url.Parse(c.host + strings.Trim(r.Path, "/"))
	if err != nil {
		return nil, fmt.Errorf("cannot parse url: %s", err)
	}

	serverUrl.Path += "/"

	req, err := http.NewRequest(method, serverUrl.String(), bytes.NewBuffer(r.Body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-WSSE", c.getWSSEHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, newClientError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response, code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (c *Client) Send(r *Request) ([]byte, error) {
	stream, err := c.SendIO(r)
	if err != nil {
		return nil, err
	}

	defer func() { _ = stream.Close() }()

	return ioutil.ReadAll(stream)
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
