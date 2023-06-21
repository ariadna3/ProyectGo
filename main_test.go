package main

import (
	"testing"
)

func Test_main(t *testing.T) {
	msg := ""
	if msg != "" {
		t.Fatalf(`Hello("") = %q, want "", error`, msg)
	}
}
