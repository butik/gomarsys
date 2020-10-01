package gomarsys

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) SendIO(r *Request) (io.ReadCloser, error) {
	args := c.Called(r)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func NewClientMock() ClientInterface {
	return &ClientMock{}
}

func (c *ClientMock) Send(r *Request) ([]byte, error) {
	args := c.Called(r)
	return args.Get(0).([]byte), args.Error(1)
}
