package httperrors

type Error interface {
	GetStatusCode() int
	error
}

type HttpError struct {
	Text   string
	Status int
}

func NewHttpError(text string, status int) HttpError {
	return HttpError{Text: text, Status: status}
}

func (e HttpError) Error() string {
	return e.Text
}

func (e HttpError) GetStatusCode() int {
	return e.Status
}
