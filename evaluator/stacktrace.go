package evaluator

import (
	"errors"

	"github.com/ocowchun/soup/lexer"
)

type StackTraceElement struct {
	//token lexer.Token
	lineNumber     int
	identifierName string
}

func (e StackTraceElement) LineNumber() int {
	return e.lineNumber
}
func (e StackTraceElement) IdentifierName() string {
	return e.identifierName
}

// how to handle runtime error with stack trace?
type RuntimeError struct {
	rawErrorMessage string
	lineNumber      int
	stackTrace      []StackTraceElement
}

func (e *RuntimeError) LineNumber() int {
	return e.lineNumber
}
func (e *RuntimeError) StackTrace() []StackTraceElement {
	return e.stackTrace
}

func (e *RuntimeError) Error() string {
	return e.rawErrorMessage
}

func newRuntimeError(err error, token lexer.Token, procedureName string) *RuntimeError {
	var prevError *RuntimeError
	if ok := errors.As(err, &prevError); ok {
		stackTrace := append(prevError.stackTrace, StackTraceElement{
			lineNumber:     prevError.lineNumber,
			identifierName: procedureName,
		})

		return &RuntimeError{
			rawErrorMessage: err.Error(),
			lineNumber:      token.Line,
			stackTrace:      stackTrace,
		}
	} else {
		return &RuntimeError{
			rawErrorMessage: err.Error(),
			lineNumber:      token.Line,
			stackTrace:      []StackTraceElement{},
		}
	}
}

//actual
//undefined identifier: `d` on line 4
//at num (line 4)
//at c (line 8)
//at b (line 11)
//at a (line 14)

// expected
//undefined identifier: `d` on line 4
//at c (line 4)
//at b (line 8)
//at a (line 11)
//at main (line 14)
