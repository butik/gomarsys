package gomarsys

import (
	"encoding/json"
	"fmt"
	"strconv"
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

const (
	ErrorCodeContactNotFound = 2008
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

type UserError struct {
	code    int
	message string
}

func (e *UserError) Error() string {
	return e.message
}

func (e *UserError) Code() int {
	return e.code
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

func (u *Users) UpdateUser(user User, keyID int) error {
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
		Method: requestPut,
		Body:   data,
	}

	if _, err := u.client.Send(r); err != nil {
		return err
	}

	return nil
}

func (u *Users) GetUserInfo(keyID int, keyValue string, fields []int) (*User, error) {
	type request struct {
		KeyID     string   `json:"keyId"`
		KeyValues []string `json:"keyValues"`
		Fields    []string `json:"fields"`
	}

	var stringFields []string
	for _, f := range fields {
		stringFields = append(stringFields, fmt.Sprintf("%d", f))
	}

	pr := &request{
		KeyID:     fmt.Sprintf("%d", keyID),
		KeyValues: []string{keyValue},
		Fields:    stringFields,
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return nil, &UserError{message: err.Error()}
	}

	r := &Request{
		Path:   "/v2/contact/getdata",
		Method: requestPost,
		Body:   data,
	}

	user := &User{}
	user.Data = make(map[int]string)

	var userData struct {
		ReplyCode int    `json:"replyCode"`
		ReplyText string `json:"replyText"`
		Data      struct {
			Errors []struct {
				Key       string `json:"key"`
				ErrorCode int    `json:"errorCode"`
				ErrorMsg  string `json:"errorMsg"`
			} `json:"errors"`
			Result interface{} `json:"result,omitempty"`
		} `json:"data"`
	}

	if response, err := u.client.Send(r); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(response, &userData); err != nil {
			return nil, &UserError{message: err.Error()}
		}
	}

	if len(userData.Data.Errors) > 0 {
		return user, &UserError{message: userData.Data.Errors[0].ErrorMsg, code: userData.Data.Errors[0].ErrorCode}
	}

	if users, ok := userData.Data.Result.([]interface{}); ok && len(userData.Data.Result.([]interface{})) > 0 {
		for key, value := range users[0].(map[string]interface{}) {
			if key == "uid" || key == "id" {
				continue
			}

			field, err := strconv.Atoi(key)
			if err != nil {
				return nil, &UserError{message: err.Error()}
			}

			if v, ok := value.(string); ok {
				user.Data[field] = v
			}
		}

		return user, nil
	}

	return nil, &UserError{message: "empty user response"}
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
