package t

import (
	"encoding/json"
	"fmt"
	"testing"
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

func TestSuccess(t *testing.T) {
	result := Success(map[string]string{
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
