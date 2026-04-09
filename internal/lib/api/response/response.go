package response

const (
	statusOK    = "OK"
	statusError = "Error"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Error(err string) Response {
	return Response{
		Status: statusError,
		Error:  err,
	}
}

func OK() Response {
	return Response{
		Status: statusOK,
	}
}
