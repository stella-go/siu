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
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db *sql.DB
)

func TestMain(m *testing.M) {
	dsn := "root:root@tcp(127.0.0.1:3306)/test?parseTime=true&collation=utf8_bin&charset=utf8"
	c, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db = c
	m.Run()
	db.Close()
}

func assertEquals(expected interface{}, actual interface{}) {
	if expected != actual {
		panic(fmt.Errorf("the expected value is %v and the actual value is %v", expected, actual))
	}
}

func TestBool(t *testing.T) {
	type S struct {
		B *Bool `json:"B,omitempty"`
	}
	// assert string
	assertEquals("{true}", fmt.Sprintf("%s", S{&Bool{true}}))
	assertEquals("{false}", fmt.Sprintf("%s", S{&Bool{false}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{true}", fmt.Sprintf("%s", &S{&Bool{true}}))
	assertEquals("&{false}", fmt.Sprintf("%s", &S{&Bool{false}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Bool{true}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"B":true}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{&Bool{false}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"B":false}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"B":true}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(true, s.B.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"B":false}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(false, s.B.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"B":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Bool)(nil), s.B)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Bool)(nil), s.B)
	}

	// assert mysql

	{
		s1 := &S{&Bool{true}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.B).Scan(&s2.B)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.B.Val, s2.B.Val)
	}
	{
		s1 := &S{&Bool{false}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.B).Scan(&s2.B)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.B.Val, s2.B.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.B).Scan(&s2.B)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.B, s2.B)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.B)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Bool)(nil), s2.B)
	}
}

func TestInt(t *testing.T) {
	type S struct {
		I *Int `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Int{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Int{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Int{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(1, s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Int{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int)(nil), s2.I)
	}
}

func TestInt8(t *testing.T) {
	type S struct {
		I *Int8 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Int8{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Int8{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Int8{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(int8(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int8)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int8)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Int8{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int8)(nil), s2.I)
	}
}

func TestInt16(t *testing.T) {
	type S struct {
		I *Int16 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Int16{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Int16{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Int16{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(int16(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int16)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int16)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Int16{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int16)(nil), s2.I)
	}
}

func TestInt32(t *testing.T) {
	type S struct {
		I *Int32 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Int32{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Int32{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Int32{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(int32(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int32)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int32)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Int32{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int32)(nil), s2.I)
	}
}

func TestRune(t *testing.T) {
	type S struct {
		I *Rune `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Rune{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Rune{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Rune{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(rune(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Rune)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Rune)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Rune{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Rune)(nil), s2.I)
	}
}

func TestInt64(t *testing.T) {
	type S struct {
		I *Int64 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Int64{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Int64{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Int64{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(int64(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int64)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int64)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Int64{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Int64)(nil), s2.I)
	}
}

func TestUint(t *testing.T) {
	type S struct {
		I *Uint `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Uint{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Uint{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Uint{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(uint(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Uint{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint)(nil), s2.I)
	}
}

func TestUint8(t *testing.T) {
	type S struct {
		I *Uint8 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Uint8{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Uint8{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Uint8{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(uint8(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint8)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint8)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Uint8{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint8)(nil), s2.I)
	}
}

func TestByte(t *testing.T) {
	type S struct {
		I *Byte `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Byte{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Byte{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Byte{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(byte(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Byte)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Byte)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Byte{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Byte)(nil), s2.I)
	}
}

func TestUint16(t *testing.T) {
	type S struct {
		I *Uint16 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Uint16{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Uint16{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Uint16{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(uint16(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint16)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint16)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Uint16{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint16)(nil), s2.I)
	}
}

func TestUint32(t *testing.T) {
	type S struct {
		I *Uint32 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Uint32{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Uint32{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Uint32{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(uint32(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint32)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint32)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Uint32{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint32)(nil), s2.I)
	}
}

func TestUint64(t *testing.T) {
	type S struct {
		I *Uint64 `json:"I,omitempty"`
	}
	// assert string
	assertEquals("{1}", fmt.Sprintf("%s", S{&Uint64{1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1}", fmt.Sprintf("%s", &S{&Uint64{1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Uint64{1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"I":1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(uint64(1), s.I.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"I":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint64)(nil), s.I)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint64)(nil), s.I)
	}

	// assert mysql
	{
		s1 := &S{&Uint64{1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I.Val, s2.I.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.I).Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.I, s2.I)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.I)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Uint64)(nil), s2.I)
	}
}

func TestFloat32(t *testing.T) {
	type S struct {
		F *Float32 `json:"F,omitempty"`
	}
	// assert string
	assertEquals("{1.1}", fmt.Sprintf("%s", S{&Float32{1.1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1.1}", fmt.Sprintf("%s", &S{&Float32{1.1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Float32{1.1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"F":1.1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"F":1.1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(float32(1.1), s.F.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"F":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float32)(nil), s.F)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float32)(nil), s.F)
	}

	// assert mysql
	{
		s1 := &S{&Float32{1.1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.F).Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.F.Val, s2.F.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.F).Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.F, s2.F)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float32)(nil), s2.F)
	}
}

func TestFloat64(t *testing.T) {
	type S struct {
		F *Float64 `json:"F,omitempty"`
	}
	// assert string
	assertEquals("{1.1}", fmt.Sprintf("%s", S{&Float64{1.1}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{1.1}", fmt.Sprintf("%s", &S{&Float64{1.1}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Float64{1.1}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"F":1.1}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"F":1.1}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(float64(1.1), s.F.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"F":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float64)(nil), s.F)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float64)(nil), s.F)
	}

	// assert mysql
	{
		s1 := &S{&Float64{1.1}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.F).Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.F.Val, s2.F.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.F).Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.F, s2.F)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.F)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Float64)(nil), s2.F)
	}
}

func TestString(t *testing.T) {
	type S struct {
		Str *String `json:"Str,omitempty"`
	}
	// assert string
	assertEquals("{abc}", fmt.Sprintf("%s", S{&String{"abc"}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{abc}", fmt.Sprintf("%s", &S{&String{"abc"}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&String{"abc"}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"Str":"abc"}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{&String{`{"obj":{},"arr":[],"i":1,"str":"abc"}`}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"Str":"{\"obj\":{},\"arr\":[],\"i\":1,\"str\":\"abc\"}"}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"Str":"abc"}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals("abc", s.Str.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"Str":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*String)(nil), s.Str)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*String)(nil), s.Str)
	}

	// assert mysql
	{
		s1 := &S{&String{"abc"}}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.Str).Scan(&s2.Str)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.Str.Val, s2.Str.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.Str).Scan(&s2.Str)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.Str, s2.Str)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.Str)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*String)(nil), s2.Str)
	}
}

func TestTime(t *testing.T) {
	now := time.Now().Round(time.Second)
	snow := now.String()
	fnow := now.Format("2006-01-02 15:04:05")
	type S struct {
		Time *Time `json:"Time,omitempty"`
	}
	// assert string
	assertEquals("{"+snow+"}", fmt.Sprintf("%s", S{&Time{now}}))
	assertEquals("{<nil>}", fmt.Sprintf("%s", S{}))

	assertEquals("&{"+snow+"}", fmt.Sprintf("%s", &S{&Time{now}}))
	assertEquals("&{<nil>}", fmt.Sprintf("%s", &S{}))

	// assert marshal json
	{
		bts, err := json.Marshal(&S{&Time{now}})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{"Time":"`+fnow+`"}`, string(bts))
	}
	{
		bts, err := json.Marshal(&S{})
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(`{}`, string(bts))
	}

	// assert unmarshal json
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"Time":"`+fnow+`"}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(now, s.Time.Val)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{"Time":null}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Time)(nil), s.Time)
	}
	{
		s := S{}
		err := json.Unmarshal([]byte(`{}`), &s)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Time)(nil), s.Time)
	}

	// assert mysql
	{
		s1 := &S{&Time{now}}
		s2 := &S{}
		err := db.QueryRow("select str_to_date(?,\"%Y-%m-%d %H:%i:%s\")", &s1.Time).Scan(&s2.Time)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.Time.Val, s2.Time.Val)
	}
	{
		s1 := &S{}
		s2 := &S{}
		err := db.QueryRow("select ?", &s1.Time).Scan(&s2.Time)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals(s1.Time, s2.Time)
	}
	{
		s2 := &S{}
		err := db.QueryRow("select NULL").Scan(&s2.Time)
		if err != nil {
			t.Fatal(err)
		}
		assertEquals((*Time)(nil), s2.Time)
	}
}
