package main

import (
	"container/heap"
	"testing"
)

func TestHeap(t *testing.T) {

	// generate the nodes
	n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)
	n2, _ := NewNodeFromIPAndPort("2001:db8::1", 4001)
	n3, _ := NewNodeFromIPAndPort("2001:db8::1", 4002)

	// Create a distance queue, put the nodes in it, and
	// establish the distance queue (heap) invariants.
	NewBoundedNodeHeap()
	nmh := make(BoundedNodeHeap, 2)

	nmh[0] = &NodeMinHeapItem{
		node:     n1,
		distance: n1.GetXorDistance(&n1),
		index:    0,
	}

	nmh[1] = &NodeMinHeapItem{
		node:     n3,
		distance: n1.GetXorDistance(&n3),
		index:    1,
	}

	heap.Init(&nmh)

	// Insert a new NodeMinHeapItem and then modify its distance.
	nmhi := &NodeMinHeapItem{
		node:     n2,
		distance: n1.GetXorDistance(&n2),
	}
	heap.Push(&nmh, nmhi)

	wanted_node := []string{
		"48a5b8b1f726b8bdf13590d01a807ccb7809f4f616340a7f6f6625e0fd84dc90",
		"6acd2ee1eb521f3fbc27e02c8b7c126864dc6270cd60ad625b243a1549253112",
		"76f7f83809e6575717560a18217d797f266b546dd5552464a92dd0909accef93",
	}

	wanted_dist := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"226896501c74a7824d1270fc91fc6ea31cd59686db54a71d34421ff5b4a1ed82",
		"3e524089fec0efeae6639ac83bfd05b45e62a09bc3612e1bc64bf57067483303",
	}

	// Take the NodeMinHeapItems out; they arrive in decreasing distance order.
	for i := 0; i < nmh.Len(); i++ {
		nmhi := heap.Pop(&nmh).(*NodeMinHeapItem)
		got_node := NodeIDToHex(nmhi.node.Node_id)
		got_dist := NodeIDToHex(nmhi.distance)

		if got_node != wanted_node[i] {
			t.Errorf("got node id: %q, wanted %q", got_node, wanted_node)
		}

		if got_dist != wanted_dist[i] {
			t.Errorf("got distance: %q, wanted %q", got_dist, wanted_dist)
		}
	}

}
