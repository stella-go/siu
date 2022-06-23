package stackerror

import (
	"fmt"
	"runtime"
	"strings"
)

type Error struct {
	Message string
	Cause   error
	Stack   []string
}

func NewError(skip int, err error) *Error {
	pc := make([]uintptr, 128)
	dep := runtime.Callers(skip, pc[:])
	stack := make([]string, 0)
	frames := runtime.CallersFrames(pc[:dep])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		stack = append(stack, fmt.Sprintf("%s\n        %s:%d", frame.Function, frame.File, frame.Line))
	}
	return &Error{
		Message: err.Error(),
		Cause:   err,
		Stack:   stack,
	}
}

func (p *Error) Unwrap() error {
	return p.Cause
}

func (p *Error) Error() string {
	stacks := strings.Join(p.Stack, "\n")
	return fmt.Sprintf("error: %s\n%s\n\ncause by: %v", p.Message, stacks, p.Cause)
}
