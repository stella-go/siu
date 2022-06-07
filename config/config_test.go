package config

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	os.Setenv("STELLA_TEST_ABC", "0x7fffffffffffffff")
	if v, ok := env.GetInt("test.abc"); ok {
		t.Log(v)
	} else {
		t.Fatal()
	}
}
