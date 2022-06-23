package t

import (
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
