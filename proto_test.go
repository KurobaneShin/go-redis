package main

import (
	"fmt"
	"testing"
)

func TestProto(t *testing.T) {
	// This is a test function.

	raw := "*3\r\n$3\r\nSET\r\n$6\r\nleader\r\n$7\r\nCharlie\r\n"

	cmd, err := parseCommand(raw)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(cmd)
}
