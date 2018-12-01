package gomarsys

import (
	"encoding/json"
	"fmt"
)

const (
	Interests = iota
	FirstName
	LastName
	EMail
	DateOfBirth
	Gender
	MaritalStatus
	Children
	Education
	Title
	Address
	City
	State
	ZIPCode
	Country
	Phone
)

type Users struct {
	client ClientInterface
}

type User struct {
	SourceID int
	Data     map[int]string
}

func NewUsers(client ClientInterface) *Users {
	return &Users{
		client: client,
	}
}

func (u *Users) Create(user User, keyID int) error {
	type request struct {
		KeyID    string              `json:"key_id"`
		Contacts []map[string]string `json:"contacts"`
	}

	pr := &request{
		KeyID: fmt.Sprintf("%d", keyID),
	}

	m := make(map[string]string)
	for key, val := range user.Data {
		m[fmt.Sprintf("%d", key)] = val
	}
	pr.Contacts = append(pr.Contacts, m)

	data, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	r := &Request{
		Path:   "/v2/contact",
		Method: requestPost,
		Body:   data,
	}

	if _, err := u.client.Send(r); err != nil {
		return err
	}

	return nil
}
