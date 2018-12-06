package gomarsys

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

type Export struct {
	client ClientInterface
}

const (
	ExportStatusScheduled  = "scheduled"
	ExportStatusInProgress = "in progress"
	ExportStatusReady      = "ready"
	ExportStatusDone       = "done"
	ExportStatusError      = "error"
)

type ExportStatus struct {
	ReplyCode int    `json:"replyCode"`
	ReplyText string `json:"replyText"`
	Data      struct {
		ID       string `json:"id"`
		Created  string `json:"created"`
		Status   string `json:"status"`
		Type     string `json:"type"`
		FileName string `json:"file_name"`
		FtpHost  string `json:"ftp_host"`
		FtpDir   string `json:"ftp_dir"`
	} `json:"data"`
}

func NewExport(client ClientInterface) *Export {
	return &Export{
		client: client,
	}
}

func (e *Export) CheckStatus(id int) (*ExportStatus, error) {
	r := &Request{
		Path:   fmt.Sprintf("/v2/export/%d", id),
		Method: requestGet,
	}

	status := &ExportStatus{}

	if response, err := e.client.Send(r); err != nil {
		return nil, err
	} else {
		err := json.Unmarshal(response, status)
		if err != nil {
			return nil, err
		}
	}

	return status, nil
}

func (e *Export) DownloadExportData(id int) ([][]string, error) {
	r := &Request{
		Path:   fmt.Sprintf("/v2/export/%d/data", id),
		Method: requestGet,
	}

	if response, err := e.client.Send(r); err != nil {
		return nil, err
	} else {
		r := csv.NewReader(strings.NewReader(string(response)))
		data, err := r.ReadAll()
		if err != nil {
			return [][]string{}, err
		}

		return data, nil
	}
}
