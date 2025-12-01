// This example demonstrates a distance queue built using the heap interface.
package main

import (
	"bytes"
)

// An NodeMinHeapItem is something we manage in a distance queue.
type NodeMinHeapItem struct {
	node    Node // Node in question
	distance NodeID    // Ordering by distance
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the NodeMinHeapItem in the heap.
}

// A NodeMinHeap implements heap.Interface and holds NodeMinHeapItems.
type NodeMinHeap []*NodeMinHeapItem

func (nmh NodeMinHeap) Len() int { return len(nmh) }

func (nmh NodeMinHeap) Less(i, j int) bool {
	// We want Pop to give us the lowest distance (nearest node).
	return bytes.Compare(nmh[i].distance, nmh[j].distance) == -1
}

func (nmh NodeMinHeap) Swap(i, j int) {
	nmh[i], nmh[j] = nmh[j], nmh[i]
	nmh[i].index = i
	nmh[j].index = j
}

func (nmh *NodeMinHeap) Push(x any) {
	n := len(*nmh)
	NodeMinHeapItem := x.(*NodeMinHeapItem)
	NodeMinHeapItem.index = n
	*nmh = append(*nmh, NodeMinHeapItem)
}

func (nmh *NodeMinHeap) Pop() any {
	old := *nmh
	n := len(old)
	NodeMinHeapItem := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the NodeMinHeapItem eventually
	NodeMinHeapItem.index = -1 // for safety
	*nmh = old[0 : n-1]
	return NodeMinHeapItem
}


