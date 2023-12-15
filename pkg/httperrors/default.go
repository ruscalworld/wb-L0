package httperrors

import "net/http"

var (
	ErrNotFound         = NewHttpError("requested resource was not found on the server", http.StatusNotFound)
	ErrMethodNotAllowed = NewHttpError("method not allowed", http.StatusMethodNotAllowed)
)
