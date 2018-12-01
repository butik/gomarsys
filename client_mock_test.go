package gomarsys

import "github.com/stretchr/testify/mock"

type ClientMock struct {
	mock.Mock
}

func NewClientMock() ClientInterface {
	return &ClientMock{}
}

func (c *ClientMock) Send(r *Request) ([]byte, error) {
	args := c.Called(r)
	return args.Get(0).([]byte), args.Error(1)
}
