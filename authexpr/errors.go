package authexpr

import (
	"fmt"
)

type Error struct {
	Msg   string
	Cause error
}

func (e Error) WithCause(cause error) Error {
	return Error{Msg: e.Msg, Cause: cause}
}

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("auth expression error: %s; cause: %s", e.Msg, e.Cause)
	}
	return fmt.Sprintf("auth expression error: %s", e.Msg)
}

func ConfigFailed(msg string, arg ...interface{}) Error {
	return Error{Msg: "configuration failure: " + fmt.Sprintf(msg, arg...)}
}

func ParseFailed(msg string, arg ...interface{}) Error {
	return Error{Msg: "expression parse error: " + fmt.Sprintf(msg, arg...)}
}

func ValidationFailed(msg string, arg ...interface{}) Error {
	return Error{Msg: "expression is invalid: " + fmt.Sprintf(msg, arg...)}
}

func EvalFailed(msg string, arg ...interface{}) Error {
	return Error{Msg: "evaluation failure: " + fmt.Sprintf(msg, arg...)}
}
