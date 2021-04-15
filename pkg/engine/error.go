package engine

import (
	"fmt"
	"net/http"
)

// HTTPErr is an error with custom HTTP status code
type HTTPErr struct {
	code int
	err  error
}

// Error creates new initialized HttpError
func Error(code int, err error) HTTPErr {
	return HTTPErr{
		code: code,
		err:  err,
	}
}

// ErrorBadRequest returns new HttpError with HTTP status code 400
func ErrorBadRequest(message string) HTTPErr {
	return Error(http.StatusBadRequest, fmt.Errorf(message))
}

// ErrorForbidden returns new HttpError with HTTP status code 403
func ErrorForbidden(message string) HTTPErr {
	return Error(http.StatusForbidden, fmt.Errorf(message))
}

// ErrorNotFound returns new HttpError with HTTP status code 404
func ErrorNotFound(message string) HTTPErr {
	return Error(http.StatusNotFound, fmt.Errorf(message))
}

// ErrorConflict returns new HttpError with HTTP status code 409
func ErrorConflict(message string) HTTPErr {
	return Error(http.StatusConflict, fmt.Errorf(message))
}

// ErrorUnprocessableEntity returns new HttpError with HTTP status code 422
func ErrorUnprocessableEntity(message string) HTTPErr {
	return Error(http.StatusUnprocessableEntity, fmt.Errorf(message))
}

// Error implements error interface
func (e HTTPErr) Error() string {
	return fmt.Sprintf("%s|||%d", e.err.Error(), e.code)
}
