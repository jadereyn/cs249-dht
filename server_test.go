package main

import (
	"math/big"
	"testing"
)

func TestLookupNodesViaBootstrap(t *testing.T) {
	// --- 1. Start bootstrap node on 127.0.0.1:9100 ---
	boot, err := NewServer("127.0.0.1", 9100)
	if err != nil {
		t.Fatalf("failed to create bootstrap node: %v", err)
	}
	// make sure we close its socket when test exits
	defer boot.Transport.conn.Close()

	// Run its RPC loop in background
	go boot.Run()

	// --- 2. Start node A on 9101 and join via bootstrap ---
	nodeA, err := NewServer("127.0.0.1", 9101)
	if err != nil {
		t.Fatalf("failed to create node A: %v", err)
	}
	defer nodeA.Transport.conn.Close()

	nodeA.PingBootstrap("127.0.0.1", 9100)

	// --- 3. Start node B on 9102 and join via bootstrap ---
	nodeB, err := NewServer("127.0.0.1", 9102)
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
