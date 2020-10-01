package gomarsys

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
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

	PlatformOriginAll             = "all"
	PlatformDefaultOriginID       = "0"
	ExportDistributionMethodLocal = "local"

	defaultCSVDelimiter = ","
)

type Users struct {
	client ClientInterface
}

type User struct {
	ID       string
	SourceID string
	Data     map[int]string
}

func (u *User) Clone() *User {
	clone := &User{
		ID:       u.ID,
		SourceID: u.SourceID,
	}

	clone.Data = make(map[int]string, len(u.Data))
	for id, value := range u.Data {
		clone.Data[id] = value
	}

	return clone
}

type BaseExportRequest struct {
	DistributionMethod  string `json:"distribution_method"`
	ContactFields       []int  `json:"contact_fields"`
	Delimiter           string `json:"delimiter"`
	AddFieldNamesHeader int    `json:"add_field_names_header"`
}

type ChangesRequest struct {
	BaseExportRequest
	Origin    string   `json:"origin"`
	TimeRange []string `json:"time_range"`
	OriginID  string   `json:"origin_id"`
}

type ContactRequest struct {
	BaseExportRequest
	ContactListID int `json:"contactlist"`
}

type SegmentRequest struct {
	BaseExportRequest
	Filter int `json:"filter"`
}

type ExportResult struct {
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

// Delete implements call to delete contact emarsys api method
// https://dev.emarsys.com/v2/contacts/delete-contact
func (u *Users) Delete(keyID int, keyValue string) error {
	keyIDString := strconv.Itoa(keyID)
	pr := map[string]string{
		"key_id":    keyIDString,
		keyIDString: keyValue,
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return &UserError{message: err.Error()}
	}

	r := &Request{
		Path:   "/v2/contact/delete",
		Method: requestPost,
		Body:   data,
	}

	var res struct {
		ReplyCode int    `json:"replyCode"`
		ReplyText string `json:"replyText"`
		Data      struct {
			Errors          json.RawMessage `json:"errors"`
			DeletedContacts uint8           `json:"deleted_contacts"`
		} `json:"data"`
	}

	if response, err := u.client.Send(r); err != nil {
		return err
	} else {
		if err := json.Unmarshal(response, &res); err != nil {
			return &UserError{message: err.Error()}
		}
	}

	err = u.handleDeleteErrors(res.Data.Errors)
	if err != nil {
		return err
	}

	if res.Data.DeletedContacts == 0 {
		return &UserError{message: fmt.Sprintf("DeletedContacts count is 0. keyID: '%d'; keyValue: '%s'", keyID, keyValue)}
	}

	return nil
}

func (u *Users) handleDeleteErrors(errorsContent json.RawMessage) error {
	// Response example:  `{"replyCode":0,"replyText":"OK","data":{"errors":{"":{"2005":"No value provided for key field: 3"}},"deleted_contacts":0}}`
	var errorSlice []struct {
		Key       string `json:"key"`
		ErrorCode int    `json:"errorCode"`
		ErrorMsg  string `json:"errorMsg"`
	}

	// Response example: `{"replyCode":0,"replyText":"OK","data":{"errors":[],"deleted_contacts":1}}`
	var errorMap map[string]map[string]string

	if err := json.Unmarshal(errorsContent, &errorSlice); err == nil {
		if len(errorSlice) > 0 {
			return &UserError{message: errorSlice[0].ErrorMsg, code: errorSlice[0].ErrorCode}
		}
	} else if err := json.Unmarshal(errorsContent, &errorMap); err == nil {
		if len(errorMap) > 0 {
			for _, errorList := range errorMap {
				for errorCodeString, errorMessage := range errorList {
					errorCode, _ := strconv.Atoi(errorCodeString)
					return &UserError{message: errorMessage, code: errorCode}
				}
			}
		}
	} else {
		return &UserError{message: fmt.Sprintf("unknown format of errors in delete response: '%s'", string(errorsContent))}
	}

	return nil
}

func (u *Users) GetUserInfo(keyID int, keyValue string, fields []int) (*User, error) {
	return u.GetUserInfoByKey(strconv.Itoa(keyID), keyValue, fields)
}

func (u *Users) GetUserInfoByKey(key, keyValue string, fields []int) (*User, error) {
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
		KeyID:     key,
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

	var userData struct {
		ReplyCode int    `json:"replyCode"`
		ReplyText string `json:"replyText"`
		Data      struct {
			Errors []struct {
				Key       string `json:"key"`
				ErrorCode int    `json:"errorCode"`
				ErrorMsg  string `json:"errorMsg"`
			} `json:"errors"`
			Result json.RawMessage `json:"result,omitempty"`
		} `json:"data"`
	}

	if response, err := u.client.Send(r); err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal(response, &userData); err != nil {
			return nil, &UserError{message: err.Error()}
		}
	}

	user := &User{}
	user.Data = make(map[int]string)

	if len(userData.Data.Errors) > 0 {
		return user, &UserError{message: userData.Data.Errors[0].ErrorMsg, code: userData.Data.Errors[0].ErrorCode}
	}

	var result []map[string]*string
	if err := json.Unmarshal(userData.Data.Result, &result); err != nil {
		return nil, &UserError{message: "empty user response"}
	}

	if len(userData.Data.Result) == 0 {
		return nil, &UserError{message: "empty user response"}
	}

	for key, value := range result[0] {
		if key == "id" && value != nil {
			user.ID = *value
		}
		if key == "uid" || key == "id" {
			continue
		}

		field, err := strconv.Atoi(key)
		if err != nil {
			return nil, &UserError{message: fmt.Errorf("cannot parse key '%s': %w", key, err).Error()}
		}

		if value != nil {
			user.Data[field] = *value
		}
	}

	return user, nil
}

func (u *Users) ListUserIDs(keyID int, keyValue string) ([]string, error) {
	query := url.Values{}
	key := strconv.Itoa(keyID)

	query.Set("return", key)
	query.Set(key, keyValue)
	query.Set("excludeempty", "false")

	path := &url.URL{
		Path:     "/v2/contact/query",
		RawQuery: query.Encode(),
	}

	r := &Request{
		Path:   path.String(),
		Method: requestGet,
	}

	var userData struct {
		ReplyCode int    `json:"replyCode"`
		ReplyText string `json:"replyText"`
		Data      struct {
			Errors []struct {
				Key       string `json:"key"`
				ErrorCode int    `json:"errorCode"`
				ErrorMsg  string `json:"errorMsg"`
			} `json:"errors"`
			Result []struct {
				ID string `json:"id"`
			} `json:"result,omitempty"`
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
		return nil, &UserError{message: userData.Data.Errors[0].ErrorMsg, code: userData.Data.Errors[0].ErrorCode}
	}

	ids := make([]string, len(userData.Data.Result))
	for i, item := range userData.Data.Result {
		ids[i] = item.ID
	}

	if len(ids) == 0 {
		return nil, &UserError{message: "empty response"}
	}

	return ids, nil
}

func (u *Users) MergeUsers(key string, sourceKeyValue, targetKeyValue string, overwriteFields []string) error {
	type request struct {
		KeyID          string            `json:"key_id"`
		SourceKeyValue string            `json:"source_key_value"`
		TargetKeyValue string            `json:"target_key_value"`
		MergeRules     map[string]string `json:"merge_rules"`
		DeleteSource   bool              `json:"delete_source"`
	}

	pr := &request{
		KeyID:          key,
		SourceKeyValue: sourceKeyValue,
		TargetKeyValue: targetKeyValue,
		MergeRules:     make(map[string]string, len(overwriteFields)),
		DeleteSource:   true,
	}

	for _, field := range overwriteFields {
		pr.MergeRules[field] = "overwrite"
	}

	data, err := json.Marshal(pr)
	if err != nil {
		return &UserError{message: err.Error()}
	}

	r := &Request{
		Path:   "/v2/contact/merge",
		Method: requestPost,
		Body:   data,
	}

	var userData struct {
		ReplyCode int    `json:"replyCode"`
		ReplyText string `json:"replyText"`
	}

	response, err := u.client.Send(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(response, &userData); err != nil {
		return &UserError{message: err.Error()}
	}

	if userData.ReplyCode != 0 {
		return &UserError{message: fmt.Sprintf("error code returned from api: '%s'", response)}
	}

	return nil
}

func (u *Users) GetSegmentLocally(segmentID int, fields []int) (*ExportResult, error) {
	return u.GetSegment(SegmentRequest{
		BaseExportRequest: BaseExportRequest{
			DistributionMethod:  ExportDistributionMethodLocal,
			ContactFields:       fields,
			AddFieldNamesHeader: 1,
			Delimiter:           defaultCSVDelimiter,
		},
		Filter: segmentID,
	})
}

func (u *Users) GetAllChangesLocally(startTime, endTime time.Time, fields []int) (*ExportResult, error) {
	return u.GetChanges(ChangesRequest{
		BaseExportRequest: BaseExportRequest{
			DistributionMethod:  ExportDistributionMethodLocal,
			ContactFields:       fields,
			AddFieldNamesHeader: 1,
			Delimiter:           defaultCSVDelimiter,
		},
		Origin: PlatformOriginAll,
		TimeRange: []string{
			startTime.Format(mysqlDateFormat),
			endTime.Format(mysqlDateFormat),
		},
		OriginID: PlatformDefaultOriginID,
	})
}

func (u *Users) GetChanges(request ChangesRequest) (*ExportResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r := &Request{
		Path:   "/v2/contact/getchanges",
		Method: requestPost,
		Body:   data,
	}

	status := &ExportResult{}

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

func (u *Users) GetContacts(request ContactRequest) (*ExportResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r := &Request{
		Path:   "/v2/email/getcontacts",
		Method: requestPost,
		Body:   data,
	}

	status := &ExportResult{}

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

func (u *Users) GetSegment(request SegmentRequest) (*ExportResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r := &Request{
		Path:   "/v2/export/filter",
		Method: requestPost,
		Body:   data,
	}

	status := &ExportResult{}

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
