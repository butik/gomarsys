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
	OptIn = 31
)

type Users struct {
	client ClientInterface
}

type User struct {
	SourceID int
	Data     map[int]string
}

type ChangesRequest struct {
	DistributionMethod  string   `json:"distribution_method"`
	Origin              string   `json:"origin"`
	TimeRange           []string `json:"time_range"`
	OriginID            string   `json:"origin_id"`
	ContactFields       []int    `json:"contact_fields"`
	Delimiter           string   `json:"delimiter"`
	AddFieldNamesHeader int      `json:"add_field_names_header"`
}

type Changes struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		ID int `json:"id"`
	} `json:"data"`
}

type ChangesResponse struct {
}

const (
	DistributionMethodLocal = "local"

	OriginAll = "all"
)

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

func (u *Users) GetChanges(request ChangesRequest) (*Changes, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r := &Request{
		Path:   "/v2/contact/getchanges",
		Method: requestPost,
		Body:   data,
	}

	status := &Changes{}

	if response, err := u.client.Send(r); err != nil {
		return nil, err
	} else {
		err := json.Unmarshal(response, status)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}
