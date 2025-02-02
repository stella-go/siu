// Copyright 2010-2025 the original author or authors.

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

var (
	NullBool       = &n.Bool{}
	NullInt        = &n.Int{}
	NullInt8       = &n.Int8{}
	NullInt16      = &n.Int16{}
	NullInt32      = &n.Int32{}
	NullRune       = NullInt32
	NullInt64      = &n.Int64{}
	NullUint       = &n.Uint{}
	NullUint8      = &n.Uint8{}
	NullByte       = NullUint8
	NullUint16     = &n.Uint16{}
	NullUint32     = &n.Uint32{}
	NullUint64     = &n.Uint64{}
	NullFloat32    = &n.Float32{}
	NullFloat64    = &n.Float64{}
	NullComplex64  = &n.Complex64{}
	NullComplex128 = &n.Complex128{}
	NullString     = &n.String{}
	NullTime       = &n.Time{}
)

func IsNull(v interface{}) bool {
	if v == nil {
		return true
	}
	switch v := v.(type) {
	case *n.Bool:
		if v == nil {
			return true
		}
		return v == NullBool
	case *n.Int:
		if v == nil {
			return true
		}
		return v == NullInt
	case *n.Int8:
		if v == nil {
			return true
		}
		return v == NullInt8
	case *n.Int16:
		if v == nil {
			return true
		}
		return v == NullInt16
	case *n.Int32:
		if v == nil {
			return true
		}
		return v == NullInt32
	/* case *n.Rune:
	return v == NullRune */
	case *n.Int64:
		if v == nil {
			return true
		}
		return v == NullInt64
	case *n.Uint:
		if v == nil {
			return true
		}
		return v == NullUint
	case *n.Uint8:
		if v == nil {
			return true
		}
		return v == NullUint8
	/* case *n.Byte:
	return v == NullByte */
	case *n.Uint16:
		if v == nil {
			return true
		}
		return v == NullUint16
	case *n.Uint32:
		if v == nil {
			return true
		}
		return v == NullUint32
	case *n.Uint64:
		if v == nil {
			return true
		}
		return v == NullUint64
	case *n.Float32:
		if v == nil {
			return true
		}
		return v == NullFloat32
	case *n.Float64:
		if v == nil {
			return true
		}
		return v == NullFloat64
	case *n.Complex64:
		if v == nil {
			return true
		}
		return v == NullComplex64
	case *n.Complex128:
		if v == nil {
			return true
		}
		return v == NullComplex128
	case *n.String:
		if v == nil {
			return true
		}
		return v == NullString
	case *n.Time:
		if v == nil {
			return true
		}
		return v == NullTime
	}
	return false
}

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

func Rune(val rune) *n.Rune {
	return &n.Rune{Val: val}
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

func Byte(val byte) *n.Byte {
	return &n.Byte{Val: val}
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

func AssertErrorNil(err error) {
	if err == nil {
		return
	}
	if serr, ok := err.(*stackerror.Error); ok {
		panic(serr)
	}
	panic(stackerror.NewError(3, err))
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
