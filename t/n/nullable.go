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

package n

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// ==================== Bool ====================
type Bool struct {
	Val bool
}

func (p Bool) MarshalJSON() ([]byte, error) {
	s := strconv.FormatBool(p.Val)
	return []byte(s), nil
}

func (p *Bool) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	p.Val = value
	return nil
}

func (p Bool) String() string {
	return strconv.FormatBool(p.Val)
}

func (p *Bool) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		if v == 0 {
			p.Val = false
		} else {
			p.Val = true
		}
	case float64:
		if v == 0 {
			p.Val = false
		} else {
			p.Val = true
		}
	case bool:
		p.Val = v
	case []byte:
		b, err := strconv.ParseBool(string(v))
		if err != nil {
			return err
		}
		p.Val = b
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		p.Val = b
	case time.Time:
		if v.IsZero() {
			p.Val = false
		} else {
			p.Val = true
		}
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to bool", value)
	}
	return nil
}

func (p Bool) Value() (driver.Value, error) {
	return p.Val, nil
}

func (p Bool) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Bool:
		return p.Val == t.Val
	case bool:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Int ====================
type Int struct {
	Val int
}

func (p Int) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(int64(p.Val), 10)
	return []byte(s), nil
}

func (p *Int) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	p.Val = int(value)
	return nil
}

func (p Int) String() string {
	return strconv.FormatInt(int64(p.Val), 10)
}

func (p *Int) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = int(v)
	case float64:
		p.Val = int(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 0)
		if err != nil {
			return err
		}
		p.Val = int(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			return err
		}
		p.Val = int(i)
	case time.Time:
		p.Val = int(v.Unix())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to int", value)
	}
	return nil
}

func (p Int) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Int) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Int:
		return p.Val == t.Val
	case int:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Int8 ====================
type Int8 struct {
	Val int8
}

func (p Int8) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(int64(p.Val), 10)
	return []byte(s), nil
}

func (p *Int8) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return err
	}
	p.Val = int8(value)
	return nil
}

func (p Int8) String() string {
	return strconv.FormatInt(int64(p.Val), 10)
}

func (p *Int8) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = int8(v)
	case float64:
		p.Val = int8(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 8)
		if err != nil {
			return err
		}
		p.Val = int8(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return err
		}
		p.Val = int8(i)
	case time.Time:
		return fmt.Errorf("can't convert type %T to int8", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to int8", value)
	}
	return nil
}

func (p Int8) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Int8) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Int8:
		return p.Val == t.Val
	case int8:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Int16 ====================
type Int16 struct {
	Val int16
}

func (p Int16) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(int64(p.Val), 10)
	return []byte(s), nil
}

func (p *Int16) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return err
	}
	p.Val = int16(value)
	return nil
}

func (p Int16) String() string {
	return strconv.FormatInt(int64(p.Val), 10)
}

func (p *Int16) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = int16(v)
	case float64:
		p.Val = int16(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 16)
		if err != nil {
			return err
		}
		p.Val = int16(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return err
		}
		p.Val = int16(i)
	case time.Time:
		return fmt.Errorf("can't convert type %T to int16", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to int16", value)
	}
	return nil
}

func (p Int16) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Int16) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Int16:
		return p.Val == t.Val
	case int16:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Int32 ====================
type Int32 struct {
	Val int32
}

func (p Int32) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(int64(p.Val), 10)
	return []byte(s), nil
}

func (p *Int32) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	p.Val = int32(value)
	return nil
}

func (p Int32) String() string {
	return strconv.FormatInt(int64(p.Val), 10)
}

func (p *Int32) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = int32(v)
	case float64:
		p.Val = int32(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 32)
		if err != nil {
			return err
		}
		p.Val = int32(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return err
		}
		p.Val = int32(i)
	case time.Time:
		p.Val = int32(v.Unix())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to int32", value)
	}
	return nil
}

func (p Int32) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Int32) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Int32:
		return p.Val == t.Val
	case int32:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Rune ====================
type Rune = Int32

// ==================== Int64 ====================
type Int64 struct {
	Val int64
}

func (p Int64) MarshalJSON() ([]byte, error) {
	s := strconv.FormatInt(p.Val, 10)
	return []byte(s), nil
}

func (p *Int64) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	p.Val = int64(value)
	return nil
}

func (p Int64) String() string {
	return strconv.FormatInt(p.Val, 10)
}

func (p *Int64) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = v
	case float64:
		p.Val = int64(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return err
		}
		p.Val = int64(i)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		p.Val = i
	case time.Time:
		p.Val = v.UnixNano()
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to int64", value)
	}
	return nil
}

func (p Int64) Value() (driver.Value, error) {
	return p.Val, nil
}

func (p Int64) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Int64:
		return p.Val == t.Val
	case int64:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Uint ====================
type Uint struct {
	Val uint
}

func (p Uint) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(p.Val), 10)
	return []byte(s), nil
}

func (p *Uint) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return err
	}
	p.Val = uint(value)
	return nil
}

func (p Uint) String() string {
	return strconv.FormatUint(uint64(p.Val), 10)
}

func (p *Uint) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = uint(v)
	case float64:
		p.Val = uint(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseUint(string(v), 10, 0)
		if err != nil {
			return err
		}
		p.Val = uint(i)
	case string:
		i, err := strconv.ParseUint(v, 10, 0)
		if err != nil {
			return err
		}
		p.Val = uint(i)
	case time.Time:
		return fmt.Errorf("can't convert type %T to uint", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to uint", value)
	}
	return nil
}

func (p Uint) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Uint) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Uint:
		return p.Val == t.Val
	case uint:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Uint8 ====================
type Uint8 struct {
	Val uint8
}

func (p Uint8) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(p.Val), 10)
	return []byte(s), nil
}

func (p *Uint8) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return err
	}
	p.Val = uint8(value)
	return nil
}

func (p Uint8) String() string {
	return strconv.FormatUint(uint64(p.Val), 10)
}

func (p *Uint8) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = uint8(v)
	case float64:
		p.Val = uint8(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseUint(string(v), 10, 8)
		if err != nil {
			return err
		}
		p.Val = uint8(i)
	case string:
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return err
		}
		p.Val = uint8(i)
	case time.Time:
		return fmt.Errorf("can't convert type %T to uint8", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to uint8", value)
	}
	return nil
}

func (p Uint8) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Uint8) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Uint8:
		return p.Val == t.Val
	case uint8:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Byte ====================
type Byte = Uint8

// ==================== Uint16 ====================
type Uint16 struct {
	Val uint16
}

func (p Uint16) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(p.Val), 10)
	return []byte(s), nil
}

func (p *Uint16) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return err
	}
	p.Val = uint16(value)
	return nil
}

func (p Uint16) String() string {
	return strconv.FormatUint(uint64(p.Val), 10)
}

func (p *Uint16) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = uint16(v)
	case float64:
		p.Val = uint16(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseUint(string(v), 10, 16)
		if err != nil {
			return err
		}
		p.Val = uint16(i)
	case string:
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return err
		}
		p.Val = uint16(i)
	case time.Time:
		return fmt.Errorf("can't convert type %T to uint16", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to uint16", value)
	}
	return nil
}

func (p Uint16) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Uint16) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Uint16:
		return p.Val == t.Val
	case uint16:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Uint32 ====================
type Uint32 struct {
	Val uint32
}

func (p Uint32) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(uint64(p.Val), 10)
	return []byte(s), nil
}

func (p *Uint32) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return err
	}
	p.Val = uint32(value)
	return nil
}

func (p Uint32) String() string {
	return strconv.FormatUint(uint64(p.Val), 10)
}

func (p *Uint32) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = uint32(v)
	case float64:
		p.Val = uint32(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseUint(string(v), 10, 32)
		if err != nil {
			return err
		}
		p.Val = uint32(i)
	case string:
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return err
		}
		p.Val = uint32(i)
	case time.Time:
		p.Val = uint32(v.Unix())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to uint32", value)
	}
	return nil
}

func (p Uint32) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Uint32) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Uint32:
		return p.Val == t.Val
	case uint32:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Uint64 ====================
type Uint64 struct {
	Val uint64
}

func (p Uint64) MarshalJSON() ([]byte, error) {
	s := strconv.FormatUint(p.Val, 10)
	return []byte(s), nil
}

func (p *Uint64) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	p.Val = uint64(value)
	return nil
}

func (p Uint64) String() string {
	return strconv.FormatUint(p.Val, 10)
}

func (p *Uint64) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = uint64(v)
	case float64:
		p.Val = uint64(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseUint(string(v), 10, 64)
		if err != nil {
			return err
		}
		p.Val = i
	case string:
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		p.Val = i
	case time.Time:
		p.Val = uint64(v.UnixNano())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to uint64", value)
	}
	return nil
}

func (p Uint64) Value() (driver.Value, error) {
	return int64(p.Val), nil
}

func (p Uint64) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Uint64:
		return p.Val == t.Val
	case uint64:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Float32 ====================
type Float32 struct {
	Val float32
}

func (p Float32) MarshalJSON() ([]byte, error) {
	s := strconv.FormatFloat(float64(p.Val), 'f', -1, 32)
	return []byte(s), nil
}

func (p *Float32) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return err
	}
	p.Val = float32(value)
	return nil
}

func (p Float32) String() string {
	return strconv.FormatFloat(float64(p.Val), 'f', -1, 32)
}

func (p *Float32) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = float32(v)
	case float64:
		p.Val = float32(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseFloat(string(v), 32)
		if err != nil {
			return err
		}
		p.Val = float32(i)
	case string:
		i, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return err
		}
		p.Val = float32(i)
	case time.Time:
		p.Val = float32(v.Unix())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to float32", value)
	}
	return nil
}

func (p Float32) Value() (driver.Value, error) {
	return float64(p.Val), nil
}

func (p Float32) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Float32:
		return p.Val == t.Val
	case float32:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Float64 ====================
type Float64 struct {
	Val float64
}

func (p Float64) MarshalJSON() ([]byte, error) {
	s := strconv.FormatFloat(float64(p.Val), 'f', -1, 32)
	return []byte(s), nil
}

func (p *Float64) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	p.Val = float64(value)
	return nil
}

func (p Float64) String() string {
	return strconv.FormatFloat(float64(p.Val), 'f', -1, 32)
}

func (p *Float64) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = float64(v)
	case float64:
		p.Val = float64(v)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		i, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return err
		}
		p.Val = i
	case string:
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		p.Val = i
	case time.Time:
		p.Val = float64(v.UnixNano())
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to float64", value)
	}
	return nil
}

func (p Float64) Value() (driver.Value, error) {
	return float64(p.Val), nil
}

func (p Float64) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Float64:
		return p.Val == t.Val
	case float64:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Complex64 ====================
type Complex64 struct {
	Val complex64
}

func (p Complex64) MarshalJSON() ([]byte, error) {
	s := strconv.FormatComplex(complex128(p.Val), 'f', -1, 64)
	return []byte(s), nil
}

func (p *Complex64) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseComplex(s, 64)
	if err != nil {
		return err
	}
	p.Val = complex64(value)
	return nil
}

func (p Complex64) String() string {
	return strconv.FormatComplex(complex128(p.Val), 'f', -1, 64)
}

func (p *Complex64) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		return fmt.Errorf("can't convert type %T to complex64", value)
	case float64:
		return fmt.Errorf("can't convert type %T to complex64", value)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		c, err := strconv.ParseComplex(string(v), 64)
		if err != nil {
			return err
		}
		p.Val = complex64(c)
	case string:
		c, err := strconv.ParseComplex(v, 64)
		if err != nil {
			return err
		}
		p.Val = complex64(c)
	case time.Time:
		return fmt.Errorf("can't convert type %T to complex64", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to complex64", value)
	}
	return nil
}

func (p Complex64) Value() (driver.Value, error) {
	return strconv.FormatComplex(complex128(p.Val), 'f', -1, 64), nil
}

func (p Complex64) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Complex64:
		return p.Val == t.Val
	case complex64:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Complex128 ====================
type Complex128 struct {
	Val complex128
}

func (p Complex128) MarshalJSON() ([]byte, error) {
	s := strconv.FormatComplex(p.Val, 'f', -1, 64)
	return []byte(s), nil
}

func (p *Complex128) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.ParseComplex(s, 64)
	if err != nil {
		return err
	}
	p.Val = value
	return nil
}

func (p Complex128) String() string {
	return strconv.FormatComplex(p.Val, 'f', -1, 64)
}

func (p *Complex128) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		return fmt.Errorf("can't convert type %T to complex128", value)
	case float64:
		return fmt.Errorf("can't convert type %T to complex128", value)
	case bool:
		if v {
			p.Val = 1
		} else {
			p.Val = 0
		}
	case []byte:
		c, err := strconv.ParseComplex(string(v), 64)
		if err != nil {
			return err
		}
		p.Val = c
	case string:
		c, err := strconv.ParseComplex(v, 64)
		if err != nil {
			return err
		}
		p.Val = c
	case time.Time:
		return fmt.Errorf("can't convert type %T to complex128", value)
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to complex128", value)
	}
	return nil
}

func (p Complex128) Value() (driver.Value, error) {
	return strconv.FormatComplex(p.Val, 'f', -1, 64), nil
}

func (p Complex128) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Complex128:
		return p.Val == t.Val
	case complex128:
		return p.Val == t
	default:
		return false
	}
}

// ==================== String ====================
type String struct {
	Val string
}

func (p String) MarshalJSON() ([]byte, error) {
	s := strconv.Quote(p.Val)
	return []byte(s), nil
}

func (p *String) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	value, err := strconv.Unquote(s)
	if err != nil {
		return err
	}
	p.Val = value
	return nil
}

func (p String) String() string {
	return p.Val
}

func (p *String) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = strconv.FormatInt(v, 10)
	case float64:
		p.Val = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		p.Val = strconv.FormatBool(v)
	case []byte:
		p.Val = string(v)
	case string:
		p.Val = v
	case time.Time:
		p.Val = v.Format("2006-01-02 15:04:05")
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to string", value)
	}
	return nil
}

func (p String) Value() (driver.Value, error) {
	return p.Val, nil
}

func (p String) Equals(o interface{}) bool {
	switch t := o.(type) {
	case String:
		return p.Val == t.Val
	case string:
		return p.Val == t
	default:
		return false
	}
}

// ==================== Time ====================
type Time struct {
	Val time.Time
}

func (p Time) MarshalJSON() ([]byte, error) {
	s := strconv.Quote(p.Val.Format("2006-01-02 15:04:05"))
	return []byte(s), nil
}

func (p *Time) UnmarshalJSON(data []byte) error {
	if data == nil {
		p = nil
		return nil
	}
	s := string(data)
	if s == "null" {
		p = nil
		return nil
	}
	tm, outerr := time.ParseInLocation("\"2006-01-02 15:04:05\"", s, time.Local)
	if outerr != nil {
		tm, err := time.ParseInLocation("\"2006-01-02\"", s, time.Local)
		if err != nil {
			return outerr
		}
		p.Val = tm
		return nil
	}
	p.Val = tm
	return nil
}

func (p Time) String() string {
	return p.Val.String()
}

func (p *Time) Scan(value interface{}) error {
	switch v := value.(type) {
	case int64:
		p.Val = time.Unix(v, 0)
	case float64:
		return fmt.Errorf("can't convert type %T to time.Time", value)
	case bool:
		return fmt.Errorf("can't convert type %T to time.Time", value)
	case []byte:
		t, err := time.ParseInLocation(time.RFC3339Nano, string(v), time.Local)
		if err != nil {
			return err
		}
		p.Val = t
	case string:
		t, err := time.ParseInLocation(time.RFC3339Nano, v, time.Local)
		if err != nil {
			return err
		}
		p.Val = t
	case time.Time:
		p.Val = v.Local()
	case nil:
		p = nil
	default:
		return fmt.Errorf("can't convert type %T to time.Time", value)
	}
	return nil
}

func (p Time) Value() (driver.Value, error) {
	return p.Val, nil
}

func (p Time) Equals(o interface{}) bool {
	switch t := o.(type) {
	case Time:
		return p.Val == t.Val
	case time.Time:
		return p.Val == t
	default:
		return false
	}
}

func (p *Time) Round(d time.Duration) *Time {
	p.Val = p.Val.Round(d)
	return p
}

type SqlNull struct{}

func (*SqlNull) Value() (driver.Value, error) {
	return nil, nil
}

var NULL = &SqlNull{}
