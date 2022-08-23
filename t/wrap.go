// Copyright 2010-2022 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package t

import (
	"fmt"
	"time"

	"github.com/stella-go/siu/t/n"
	"github.com/stella-go/siu/t/stackerror"
)

func Int(val int) *n.Int {
	return &n.Int{Val: val}
}

func Int8(val int8) *n.Int8 {
	return &n.Int8{Val: val}
}

func Int16(val int16) *n.Int16 {
	return &n.Int16{Val: val}
}

func Int32(val int32) *n.Int32 {
	return &n.Int32{Val: val}
}

func Int64(val int64) *n.Int64 {
	return &n.Int64{Val: val}
}

func Uint(val uint) *n.Uint {
	return &n.Uint{Val: val}
}

func Uint8(val uint8) *n.Uint8 {
	return &n.Uint8{Val: val}
}

func Uint16(val uint16) *n.Uint16 {
	return &n.Uint16{Val: val}
}

func Uint32(val uint32) *n.Uint32 {
	return &n.Uint32{Val: val}
}

func Uint64(val uint64) *n.Uint64 {
	return &n.Uint64{Val: val}
}

func Float32(val float32) *n.Float32 {
	return &n.Float32{Val: val}
}

func Float64(val float64) *n.Float64 {
	return &n.Float64{Val: val}
}

func Complex64(val complex64) *n.Complex64 {
	return &n.Complex64{Val: val}
}

func Complex128(val complex128) *n.Complex128 {
	return &n.Complex128{Val: val}
}

func String(val string) *n.String {
	return &n.String{Val: val}
}

func Time(val time.Time) *n.Time {
	return &n.Time{Val: val}
}

func Error(err error) *stackerror.Error {
	if serr, ok := err.(*stackerror.Error); ok {
		return serr
	}
	return stackerror.NewError(3, err)
}

func Errorf(format string, a ...interface{}) *stackerror.Error {
	err := fmt.Errorf(format, a...)
	return stackerror.NewError(3, err)
}

func Success() *ResultBean[any] {
	return &ResultBean[any]{Code: 200, Message: "success"}
}

func SuccessWith[T any](data T) *ResultBean[T] {
	return &ResultBean[T]{Code: 200, Message: "success", Data: data}
}

func Fail() *ResultBean[any] {
	return &ResultBean[any]{Code: 500, Message: "failed"}
}

func FailWith(code int, message string) *ResultBean[any] {
	return &ResultBean[any]{Code: code, Message: message}
}
