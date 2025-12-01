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

func TestDistanceToSelf(t *testing.T) {

    n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)

    got := NodeIDToHex(n1.GetXorDistance(&n1))
    want := "0000000000000000000000000000000000000000000000000000000000000000"

    if got != want {
        t.Errorf("got %q, wanted %q", got, want)
    }
}

func TestDistanceSymmetry(t *testing.T) {

    n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)
	n2, _ := NewNodeFromIPAndPort("2001:db8::1", 4001)

    dist_fw := NodeIDToHex(n1.GetXorDistance(&n2))
    dist_rev := NodeIDToHex(n2.GetXorDistance(&n1))

    if dist_fw != dist_rev {
        t.Errorf("forward distance: %q, reverse distance: %q", dist_fw, dist_rev)
    }
}

func TestDistanceGeneral(t *testing.T) {

    n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)
	n2, _ := NewNodeFromIPAndPort("2001:db8::1", 4001)
	n3, _ := NewNodeFromIPAndPort("2001:db8::1", 4002)

    got_13 :=  NodeIDToHex(n1.GetXorDistance(&n3))
    want_13 := "226896501c74a7824d1270fc91fc6ea31cd59686db54a71d34421ff5b4a1ed82"

    if got_13 != want_13 {
        t.Errorf("got %q, wanted %q", got_13, want_13)
    }

	got_23 := NodeIDToHex(n2.GetXorDistance(&n3))
	want_23 := "1c3ad6d9e2b44868ab71ea34aa016b1742b7361d18358906f209ea85d3e9de81"

	if got_23 != want_23 {
        t.Errorf("got %q, wanted %q", got_23, want_23)
    }

}