package main

import (
	"testing"
)

func TestNewFromIPAndPort(t *testing.T) {

	n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)
    got :=  n1.HexID()
    want := "48a5b8b1f726b8bdf13590d01a807ccb7809f4f616340a7f6f6625e0fd84dc90"

    if got != want {
        t.Errorf("got %q, wanted %q", got, want)
    }

	got_a := n1.ip_addr
	want_a := "192.0.2.10"

	if got != want {
        t.Errorf("got %q, wanted %q", got_a, want_a)
    }

	got_p := n1.port
	want_p := 4001

	if got != want {
        t.Errorf("got %q, wanted %q", got_p, want_p)
    }
}