package main

import (
	"math/big"
	"testing"
	"time"
)

func TestNewFromIPAndport(t *testing.T) {

	n1, _ := NewNodeFromIPAndport("192.0.2.10", 4001)
	got := n1.HexID()
	want := "48a5b8b1f726b8bdf13590d01a807ccb7809f4f616340a7f6f6625e0fd84dc90"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

	got_a := n1.ipAddr
	want_a := "192.0.2.10"

	if got_a != want_a {
		t.Errorf("got %q, wanted %q", got_a, want_a)
	}

	got_p := n1.port
	want_p := 4001

	if got_p != want_p {
		t.Errorf("got %q, wanted %q", got_p, want_p)
	}
}

func TestDistanceToSelf(t *testing.T) {

	n1, _ := NewNodeFromIPAndport("192.0.2.10", 4001)

	got := NodeIDToHex(n1.GetXorDistance(&n1))
	want := "0000000000000000000000000000000000000000000000000000000000000000"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestDistanceSymmetry(t *testing.T) {

	n1, _ := NewNodeFromIPAndport("192.0.2.10", 4001)
	n2, _ := NewNodeFromIPAndport("2001:db8::1", 4001)

	dist_fw := NodeIDToHex(n1.GetXorDistance(&n2))
	dist_rev := NodeIDToHex(n2.GetXorDistance(&n1))

	if dist_fw != dist_rev {
		t.Errorf("forward distance: %q, reverse distance: %q", dist_fw, dist_rev)
	}
}

func TestDistanceGeneral(t *testing.T) {

	n1, _ := NewNodeFromIPAndport("192.0.2.10", 4001)
	n2, _ := NewNodeFromIPAndport("2001:db8::1", 4001)
	n3, _ := NewNodeFromIPAndport("2001:db8::1", 4002)

	got_13 := NodeIDToHex(n1.GetXorDistance(&n3))
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

// small helper so we don't repeat sleep logic
func waitForRouting() {
	time.Sleep(300 * time.Millisecond)
}

func TestLookupNodesViaBootstrap(t *testing.T) {
	// --- 1. Start bootstrap node on 127.0.0.1:9100 ---
	boot, err := NewLocalNode("127.0.0.1", 9100)
	if err != nil {
		t.Fatalf("failed to create bootstrap node: %v", err)
	}
	// make sure we close its socket when test exits
	defer boot.Transport.conn.Close()

	// Run its RPC loop in background
	go boot.Run()

	// --- 2. Start node A on 9101 and join via bootstrap ---
	nodeA, err := NewLocalNode("127.0.0.1", 9101)
	if err != nil {
		t.Fatalf("failed to create node A: %v", err)
	}
	defer nodeA.Transport.conn.Close()

	nodeA.PingBootstrap("127.0.0.1", 9100)

	// --- 3. Start node B on 9102 and join via bootstrap ---
	nodeB, err := NewLocalNode("127.0.0.1", 9102)
	if err != nil {
		t.Fatalf("failed to create node B: %v", err)
	}
	defer nodeB.Transport.conn.Close()

	nodeB.PingBootstrap("127.0.0.1", 9100)

	// Give the network a tiny bit of time so the Ping / Pong
	// and routing table updates can complete.
	waitForRouting()

	// --- 4. Run a Kademlia lookup from B for A's ID ---
	targetID := new(big.Int).Set(nodeA.Self.nodeID)

	neighbors, err := nodeB.LookupNodes(targetID)
	if err != nil {
		t.Fatalf("LookupNodes error: %v", err)
	}

	if len(neighbors) == 0 {
		t.Fatalf("LookupNodes returned no neighbors")
	}

	// --- 5. Check that A is among the neighbors ---
	foundA := false
	targetHex := nodeA.Self.HexID()
	for _, n := range neighbors {
		if n.HexID() == targetHex {
			foundA = true
			break
		}
	}

	if !foundA {
		t.Fatalf("expected to find node A (id=%s) among neighbors, got: %+v", targetHex, neighbors)
	}
}
