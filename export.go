package gomarsys

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type Export struct {
	client ClientInterface
}

const (
	ExportStatusScheduled  = "scheduled"
	ExportStatusInProgress = "in_progress"
	ExportStatusReady      = "ready"
	ExportStatusDone       = "done"
	ExportStatusError      = "error"

	bufferSize = 512

	emarsysUpdateStatusPeriod = time.Second * 10
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

func (e *Export) WaitExportComplete(ctx context.Context, jobID int) (*ExportStatus, error) {
	var (
		status *ExportStatus
		err    error
	)

	t := time.NewTicker(emarsysUpdateStatusPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
			break
		}

		status, err = e.CheckStatus(jobID)
		if err != nil {
			return nil, err
		}

		if status.Data.Status == ExportStatusDone || status.Data.Status == ExportStatusError {
			break
		}
	}

	if status == nil {
		return nil, fmt.Errorf("unknown export status, job id: %d", jobID)
	}

	return status, nil
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
			return nil, err
		}

		return data, nil
	}
}

func (e *Export) copyIO(source io.Reader, destination io.Writer) error {
	buf := make([]byte, bufferSize)

	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Export) DownloadExportToIO(id int, stream io.Writer) error {
	r := &Request{
		Path:   fmt.Sprintf("/v2/export/%d/data", id),
		Method: requestGet,
	}

	if responseStream, err := e.client.SendIO(r); err != nil {
		return err
	} else {
		defer func() { _ = responseStream.Close() }()

		return e.copyIO(responseStream, stream)
	}
}
