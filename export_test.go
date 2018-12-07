package gomarsys

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExport_CheckStatus(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/export/1")
		assert.Equal(t, req.Method, RequestMethod(requestGet))
	}).Return([]byte(`{"replyCode": 0,"replyText":"ok","data":{"id":"123","status":"done"}}`), nil)

	export := NewExport(client)
	exportStatus, err := export.CheckStatus(1)
	require.NoError(t, err)
	assert.Equal(t, exportStatus.ReplyCode, 0)
	assert.Equal(t, exportStatus.ReplyText, "ok")
	assert.Equal(t, exportStatus.Data.ID, "123")
	assert.Equal(t, exportStatus.Data.Status, ExportStatusDone)
}

func TestExport_DownloadExportData(t *testing.T) {
	client := NewClientMock()
	client.(*ClientMock).On("Send", mock.Anything).Run(func(args mock.Arguments) {
		req := args.Get(0).(*Request)

		assert.Equal(t, req.Path, "/v2/export/1/data")
		assert.Equal(t, req.Method, RequestMethod(requestGet))
	}).Return([]byte(`123,10-1000,True`), nil)

	export := NewExport(client)
	exportStatus, err := export.DownloadExportData(1)
	require.NoError(t, err)
	assert.Equal(t, exportStatus[0][0], "123")
	assert.Equal(t, exportStatus[0][1], "10-1000")
	assert.Equal(t, exportStatus[0][2], "True")
}
