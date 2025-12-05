package main

import (
	"testing"
	"math/big"
	"fmt"
)

func TestRouter(t *testing.T) {
	fmt.Println("testing router #######3")
	our_node := NewNodeFromInt(1)
	router := NewRouter(our_node)

	contact := NewNodeFromInt(2)
	router.AddContact(contact)

	contact = NewNodeFromInt(3)
	router.AddContact(contact)

	contact = NewNodeFromInt(4)
	router.AddContact(contact)

	if len(router.buckets) != 2 {
		t.Errorf("got %q, wanted %q", len(router.buckets), 2)
	}

	fb := router.buckets[1]

	if len(fb.GetNodes()) != 3 {
		t.Errorf("got %q, wanted %q", len(fb.GetNodes()), 3)
	}
}

func TestTraversal(t *testing.T) {

	var nodes [10]Node
	for i := 0; i < 10; i++ {
		newNode := NewNodeFromInt(int64(i))
		nodes[i] = newNode
	}

	var buckets []*KBucket

	for i := 0; i < 5; i++ {
		bucket := NewKBucket(big.NewInt(int64(2 * i)), big.NewInt(int64(2 * i + 1)))
		bucket.AddNode(nodes[2 * i])
		bucket.AddNode(nodes[2 * i + 1])
		buckets = append(buckets, &bucket)
	}

	our_node := NewNodeFromInt(20)
	router := NewRouter(our_node)

	// replace with test buckets
	router.buckets = buckets

	expected_nodes := []Node {nodes[5], nodes[4], nodes[3], nodes[2], nodes[7], nodes[6], nodes[1], nodes[0], nodes[9], nodes[8]}

	start_node := nodes[4]
	traverser := NewTraversal(&router, start_node)

	neighbor, isComplete := traverser.Next()
	index := 0;

	for !isComplete {
		
		if neighbor.HexID() != expected_nodes[index].HexID() {
			t.Errorf("got %q, wanted %q", neighbor.HexID(), expected_nodes[index].HexID())
		}

		neighbor, isComplete = traverser.Next()
		index++
	}

}