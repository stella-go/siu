// Copyright 2010-2024 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// in general, skip is 3
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
