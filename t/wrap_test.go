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
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stella-go/siu/t/n"
)

func TestError(t *testing.T) {
	err := err0()
	fmt.Printf("%s\n", err)
}

func err0() error {
	return err1()
}
func err1() error {
	return err2()
}
func err2() error {
	err := fmt.Errorf("this is an error")
	fmt.Printf("%v\n", err)
	return Error(err)
}

func TestErrorf(t *testing.T) {
	err := Errorf("Found an error: %v", errors.New("this is an error"))
	fmt.Printf("%s\n", err)
}

func TestSuccess(t *testing.T) {
	result := SuccessWith(map[string]string{
		"foo": "B",
		"bar": "A",
	})
	t.Log(result)
	bts, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bts))
}

func TestFail(t *testing.T) {
	result := Fail()
	t.Log(result)
	bts, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bts))
}

func TestFailWith(t *testing.T) {
	result := FailWith(500, "err")
	t.Log(result)
	bts, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bts))
}

func TestIsNull(t *testing.T) {
	if !IsNull(nil) {
		t.Error("Expected nil to be null")
	}

	var nullInt *n.Int
	if !IsNull(nullInt) {
		t.Error("Expected *n.Int nil to be null")
	}

	notNullInt := Int(5)
	if IsNull(notNullInt) {
		t.Error("Expected non-nil *n.Int to not be null")
	}

	if !IsNull(NullInt) {
		t.Error("Expected NullInt nil to not be null")
	}

	notNullString := String("hello")
	if IsNull(notNullString) {
		t.Error("Expected non-nil *n.String to not be null")
	}
}
