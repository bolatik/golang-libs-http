package http

import (
	"fmt"
	"net/http"
	"strconv"
	"encoding/json"
)

type (
	BaseError interface {
		Headers() map[string]string

		Error() string

		Response() interface{}

		StatusCode() int
	}

	Error struct {
		System           string `json:"system,omitempty"`
		Status           int    `json:"status,omitempty"`
		Series           int    `json:"series,omitempty"`
		Code             string `json:"code,omitempty"`
		Message          string `json:"message,omitempty"`
		DeveloperMessage string `json:"developerMessage,omitempty"`
		MoreInfo         string `json:"moreInfo,omitempty"`
	}

	ErrorSystem struct {
		System string `json:"system,omitempty"`
		Series int    `json:"series,omitempty"`
	}
)

func (e Error) Headers() map[string]string {
	return nil
}

func (Error) SetHeader(key, value string) {}

func (Error) GetHeader(key string) string {
	return ""
}

func (e Error) Response() interface{} {
	return e
}

func (e Error) Error() string {
	return fmt.Sprintf("system: %s; status: %d; code: %s; message: %s; moreInfo: %s", e.System, e.Status, e.Code, e.Message, e.MoreInfo)
}

func (e Error) StatusCode() int {
	return e.Status
}

func NewErrorSystem(system string, series int) *ErrorSystem {
	return &ErrorSystem{
		System: system,
		Series: series,
	}
}

func (errSys *ErrorSystem) NewError(status, code int, messages ...string) *Error {
	var message, devMessage string

	switch len(messages) {
	case 2:
		message, devMessage = messages[0], messages[1]
	case 1:
		message = messages[0]
	}

	c := errSys.System + "." + strconv.Itoa(status) + strconv.Itoa(errSys.Series) + strconv.Itoa(code)

	return &Error{
		System:           errSys.System,
		Status:           status,
		Series:           errSys.Series,
		Code:             c,
		Message:          message,
		DeveloperMessage: devMessage,
		MoreInfo:         "docs/" + c,
	}
}

func (errSys *ErrorSystem) FromError(err BaseError) BaseError {
	derr, ok := err.(Error)

	if ok {
		c, _ := strconv.Atoi(derr.Code)
		return errSys.NewError(err.StatusCode(), c, derr.Message, derr.DeveloperMessage)
	}

	derr1, ok := err.(*Error)
	if ok {
		c, _ := strconv.Atoi(derr1.Code)
		return errSys.NewError(err.StatusCode(), c, derr1.Message, derr.DeveloperMessage)
	}

	return nil
}

func (errSys *ErrorSystem) BadRequest(code int, messages ...string) *Error {
	return errSys.NewError(http.StatusBadRequest, code, messages...)
}

func (errSys *ErrorSystem) InternalServerError(code int, messages ...string) *Error {
	return errSys.NewError(http.StatusInternalServerError, code, messages...)
}

func (errSys *ErrorSystem) NotFound(code int, messages ...string) *Error {
	return errSys.NewError(http.StatusNotFound, code, messages...)
}

func (errSys *ErrorSystem) Forbidden(code int, messages ...string) *Error {
	return errSys.NewError(http.StatusForbidden, code, messages...)
}

func IfJsonError(data []byte) *Error {
	er := &Error{}
	err := json.Unmarshal(data, er)
	if err != nil {
		return nil
	}
	isErr := er.Series != 0 && er.Message != "" && er.Code != "" && er.System != ""
	if !isErr {
		return nil
	}
	return er
}
