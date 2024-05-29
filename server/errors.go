package server

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type PCError struct {
	Code         int
	FullPage     bool
	TemplateData map[string]interface{}
	Message      string
}

func ErrorMessage(code int, message string) PCError {
	return PCError{
		Code:     code,
		Message:  message,
		FullPage: false,
	}
}

func (e PCError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Internal Server Error"
}

func ErrorHandler(c echo.Context, err error) {
	if err == nil || c.Response().Committed {
		return
	}

	var e PCError = PCError{}

	if !errors.As(err, &e) {
		e = ErrorMessage(500, err.Error())
	}

	c.HTML(e.Code, e.Error())
}
