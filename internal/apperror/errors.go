package apperror

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	ExitConfig       = 10
	ExitInput        = 11
	ExitAPI          = 12
	ExitRPC          = 13
	ExitExecution    = 14
	ExitVerification = 15
)

type Error struct {
	Stage   string `json:"stage"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *Error) ExitCode() int {
	if e == nil || e.Code == 0 {
		return 1
	}
	return e.Code
}

func New(stage string, code int, message string) error {
	return &Error{
		Stage:   stage,
		Status:  "error",
		Message: message,
		Code:    code,
	}
}

func Wrap(stage string, code int, err error) error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return err
	}
	return &Error{
		Stage:   stage,
		Status:  "error",
		Message: err.Error(),
		Code:    code,
		Err:     err,
	}
}

func Classify(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	message := err.Error()
	lower := strings.ToLower(message)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return &Error{Stage: "rpc", Status: "error", Message: message, Code: ExitRPC, Err: err}
	case strings.Contains(lower, "api error"):
		return &Error{Stage: "api", Status: "error", Message: message, Code: ExitAPI, Err: err}
	case strings.Contains(lower, "rpc") ||
		strings.Contains(lower, "dial") ||
		strings.Contains(lower, "nonce") ||
		strings.Contains(lower, "receipt"):
		return &Error{Stage: "rpc", Status: "error", Message: message, Code: ExitRPC, Err: err}
	case strings.Contains(lower, "timed out waiting for portfolio") ||
		strings.Contains(lower, "position detected: no"):
		return &Error{Stage: "verification", Status: "error", Message: message, Code: ExitVerification, Err: err}
	case strings.Contains(lower, "required") ||
		strings.Contains(lower, "unknown") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "mutually exclusive") ||
		strings.Contains(lower, "must be one of") ||
		strings.Contains(lower, "invalid"):
		return &Error{Stage: "input", Status: "error", Message: message, Code: ExitInput, Err: err}
	default:
		return &Error{Stage: "execution", Status: "error", Message: message, Code: ExitExecution, Err: err}
	}
}

func JSONPayload(err error) map[string]any {
	appErr := Classify(err)
	return map[string]any{
		"stage":   appErr.Stage,
		"status":  appErr.Status,
		"message": appErr.Message,
		"code":    appErr.Code,
	}
}

func Formatf(stage string, code int, format string, args ...any) error {
	return &Error{
		Stage:   stage,
		Status:  "error",
		Message: fmt.Sprintf(format, args...),
		Code:    code,
	}
}
