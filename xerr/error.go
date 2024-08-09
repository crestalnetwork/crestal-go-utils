package xerr

import (
	"errors"
	"fmt"
)

// ServerError always the same
var ServerError = New(500, "ServerError",
	"There was an issue on the server side. Please report to us or try again later.")

// Error custom struct
type Error struct {
	err     error // support the Unwrap interface
	code    int
	Key     string `json:"error"`
	Message string `json:"message"`
}

// New Error
func New(code int, key string, msg string) *Error {
	return &Error{
		code:    code,
		Key:     key,
		Message: msg,
	}
}

// Newf create an Error use format
func Newf(code int, key string, format string, a ...interface{}) *Error {
	err := fmt.Errorf(format, a...)
	return &Error{
		err:     err,
		code:    code,
		Key:     key,
		Message: err.Error(),
	}
}

// Error makes it compatible with `error` interface.
func (e *Error) Error() string {
	return e.Message
}

// StatusCode is http status code
func (e *Error) StatusCode() int {
	return e.code
}

// Unwrap support the Unwrap interface
func (e *Error) Unwrap() error {
	return e.err
}

// Is err the instance of Error,and has <key>?
func Is(err error, key string) bool {
	src, ok := As(err)
	if !ok {
		return false
	}
	if src.Key == key {
		return true
	}
	return false
}

// IsCode check if the status code is <code>
func IsCode(err error, code int) bool {
	src, ok := As(err)
	if !ok {
		return false
	}
	if src.code == code {
		return true
	}
	return false
}

func As(err error) (*Error, bool) {
	e := new(Error)
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
