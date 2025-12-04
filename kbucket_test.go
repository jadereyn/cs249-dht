package main

import (
	"testing"
	"math/big"
)

func TestSplit(t *testing.T) {

	bucket := NewKBucket(big.NewInt(0), big.NewInt(KSIZE * 2))
	n1 := NewNodeFromInt(KSIZE)
	n2 := NewNodeFromInt(KSIZE + 1)

	bucket.AddNode(n1)
	bucket.AddNode(n2)

	first, second := bucket.Split()

	got := first.Len()
	want := 1

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

	got = second.Len()
	want = 1

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

	got2 := first.range_lower
	want2 := big.NewInt(0)

	if got2.Cmp(want2) != 0 {
		t.Errorf("got %q, wanted %q", got2, want2)
	}

	got2 = first.range_upper
	want2 = big.NewInt(KSIZE)

	if got2.Cmp(want2) != 0 {
		t.Errorf("got %q, wanted %q", got2, want2)
	}

	got2 = second.range_lower
	want2 = big.NewInt(KSIZE + 1)

	if got2.Cmp(want2) != 0 {
		t.Errorf("got %q, wanted %q", got2, want2)
	}

	got2 = second.range_upper
	want2 = big.NewInt(KSIZE * 2)

	if got2.Cmp(want2) != 0 {
		t.Errorf("got %q, wanted %q", got2, want2)
	}

}

func TestSplitNoOverlap(t *testing.T) {
	upper := big.NewInt(1)
	upper.Lsh(upper, NODE_ID_BIT_SIZE)
	bucket := NewKBucket(big.NewInt(0), upper)
	left, right := bucket.Split()
	
	got := left.range_upper
	want := right.range_lower

	res := want.Sub(want, got)

	if res.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("got %q, wanted %q", got, want)
	}

}

func TestAddNode(t *testing.T) {
	bucket := NewKBucket(big.NewInt(0), big.NewInt(KSIZE * 5))

	for i := 0; i < KSIZE; i++ {
		newNode := NewNodeFromInt(int64(i))
		got := bucket.AddNode(newNode)
		want := true

		if got != want {
			t.Errorf("got %t, wanted %t. i = %q", got, want, i)
		}
	}

	newNode := NewNodeFromInt(KSIZE)
	got := bucket.AddNode(newNode)
	want := false

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}
}

func TestDoubleAddNode(t *testing.T) {
	bucket := NewKBucket(big.NewInt(0), big.NewInt(KSIZE * 5))

	var nodelist [KSIZE]Node
	for i := 0; i < KSIZE; i++ {
		newNode := NewNodeFromInt(int64(i))
		nodelist[i] = newNode
	}

	for i, newNode := range nodelist {
		got := bucket.AddNode(newNode)
		want := true

		if got != want {
			t.Errorf("got %t, wanted %t. i = %q", got, want, i)
		}
	}

	for index, node := range bucket.GetNodes() {
		if node != nodelist[index] {
			t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index].HexID(), index)
		}
	}

	// add first node in nodelist again
	got := bucket.AddNode(nodelist[0])
	want := true

	if got != want {
		t.Errorf("got %t, wanted %t", got, want)
	}

	for index, node := range bucket.GetNodes() {
		if index != KSIZE-1 {
			if node != nodelist[index+1] {
				t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index+1].HexID(), index)
			}
		} else if index == KSIZE -1 {
			if node != nodelist[0] {
				t.Errorf("got %q %q %q, wanted %q %q %q. index = %q", node.HexID(), node.ipAddr, node.port, nodelist[0].HexID(), nodelist[0].ipAddr, nodelist[0].port, index)
			}
		}
	}
}

func TestRemoveNode(t *testing.T) {
	bucket := NewKBucket(big.NewInt(0), big.NewInt(KSIZE + 5))

	var nodelist [KSIZE + 5]Node
	for i := 0; i < KSIZE + 5; i++ {
		newNode := NewNodeFromInt(int64(i))
		nodelist[i] = newNode
	}

	for _, newNode := range nodelist {
		bucket.AddNode(newNode)
	}

	for index, node := range bucket.GetNodes() {
		if node != nodelist[index] {
			t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index].HexID(), index)
		}
	}	

	for index, repl_node := range bucket.GetReplacementNodes() {
		if repl_node != nodelist[index+KSIZE] {
			t.Errorf("got %q, wanted %q. index = %q", repl_node.HexID(), nodelist[index+KSIZE].HexID(), index+KSIZE)
		}
	}

	// remove last node that was added (newest node)
	// same invariant should be true as above
	bucket.RemoveNode(nodelist[KSIZE+4])
	for index, node := range bucket.GetNodes() {
		if node != nodelist[index] {
			t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index].HexID(), index)
		}
	}	

	for index, repl_node := range bucket.GetReplacementNodes() {
		if repl_node != nodelist[index+KSIZE] {
			t.Errorf("got %q, wanted %q. index = %q", repl_node.HexID(), nodelist[index+KSIZE].HexID(), index+KSIZE)
		}
	}

	// remove first node that was added (oldest node)
	// most recent replacement node is added to the node list
	// and removed from the replacement node list
	bucket.RemoveNode(nodelist[0])
	for index, node := range bucket.GetNodes() {
		if index != KSIZE -1 {
			if node != nodelist[index+1] {
				t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index].HexID(), index)
			}
		} else {
			if node != nodelist[index+4] {
				t.Errorf("got %q, wanted %q. index = %q", node.HexID(), nodelist[index+4].HexID(), index)
			}
		}

	}	

	for index, repl_node := range bucket.GetReplacementNodes() {
		if repl_node != nodelist[index+KSIZE] {
			t.Errorf("got %q, wanted %q. index = %q", repl_node.HexID(), nodelist[index+KSIZE].HexID(), index+KSIZE)
		}
	}

	if len(bucket.GetReplacementNodes()) != REPLACEMENT_FACTOR - 2 { // removed 2 nodes
		t.Errorf("got %q, wanted %q", len(bucket.GetReplacementNodes()), REPLACEMENT_FACTOR - 2)
	}

}

func TestInRange(t *testing.T) {
	bucket := NewKBucket(big.NewInt(0), big.NewInt(10))

	n0 := NewNodeFromInt(0)
	n5 := NewNodeFromInt(5)
	n10 := NewNodeFromInt(10)
	n11 := NewNodeFromInt(11)

	got := bucket.HasInRange(n0.nodeID)

	if got != true {
		t.Errorf("got %t, wanted %t", got, true)
	}

	got = bucket.HasInRange(n5.nodeID)

	if got != true {
		t.Errorf("got %t, wanted %t", got, true)
	}

	got = bucket.HasInRange(n10.nodeID)

	if got != true {
		t.Errorf("got %t, wanted %t", got, true)
	}

	got = bucket.HasInRange(n11.nodeID)

	if got != false {
		t.Errorf("got %t, wanted %t", got, false)
	}

}