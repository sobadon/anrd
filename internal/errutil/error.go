package errutil

var (
	ErrHTTPRequest     = NewInternalError("http request error")
	ErrJSONDecode      = NewInternalError("json decode error")
	ErrTimeParse       = NewInternalError("time parse error")
	ErrGetProgramNotOK = NewInternalError("http get program status code not ok")
	ErrDatabaseOpen    = NewInternalError("database open error")
	ErrDatabaseQuery   = NewInternalError("database query error")
	ErrDatabaseScan    = NewInternalError("database scan error")
	ErrDatabasePrepare = NewInternalError("database prepare error")
	ErrFfmpeg          = NewInternalError("ffmpeg error")
	ErrScheduler       = NewInternalError("scheduler error")
	// 分類できない系
	ErrInternal = NewInternalError("internal something error")
)
