package structs

type ErrorMessage struct {
	Error string `json:"error"`
}

func NewErrorMessage(err string) *ErrorMessage {
	return &ErrorMessage{
		Error: err,
	}
}
