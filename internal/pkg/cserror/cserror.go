package cserror

import "net/http"

type CsError struct {
}

func New() *CsError {
	return &CsError{}
}

// The clientError helper sends a specific status code and corresponding descri
// to the user. We'll use this later in the book to send responses like 400 "Bad Request"
// when there's a problem with the request that the user sent.
func (e *CsError) ClientError(w http.ResponseWriter, status int, err error) {
	// app.logger.Err(err).Msg("clientError")
	http.Error(w, http.StatusText(status), status)
}

// The ServerError helper writes an error message and stack trace to the errorLo
// then sends a generic 500 Internal Server Error response to the user.
func (e *CsError) ServerError(w http.ResponseWriter, err error) {
	// trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// app.logger.Err(err).Msg("")

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
