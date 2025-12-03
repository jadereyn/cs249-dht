// This example demonstrates a distance queue built using the heap interface.
package main

import (
	"bytes"
	"container/heap"
	"sort"
)

// An NodeMinHeapItem is something we manage in a distance queue.
type NodeMinHeapItem struct {
	node Node // Node in question
	// TODO: change to *big.Int for arbitrary-length IDs
	// distance *big.Int // Ordering by distance
	distance NodeID // Ordering by distance
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the NodeMinHeapItem in the heap.
}

// A NodeMinHeap implements heap.Interface and holds NodeMinHeapItems.
// type NodeMinHeap []*NodeMinHeapItem

// func (nmh NodeMinHeap) Len() int { return len(nmh) }

// func (nmh NodeMinHeap) Less(i, j int) bool {
// 	// We want Pop to give us the lowest distance (nearest node).
// 	return bytes.Compare(nmh[i].distance, nmh[j].distance) == -1
// }

// TODO : change to *big.Int for arbitrary-length IDs
// A NodeMinHeap implements heap.Interface and holds NodeMinHeapItems.
// func (nmh NodeMinHeap) Less(i, j int) bool {
//     return nmh[i].Distance.Cmp(nmh[j].Distance) < 0
// }

// func (nmh NodeMinHeap) Swap(i, j int) {
// 	nmh[i], nmh[j] = nmh[j], nmh[i]
// 	nmh[i].index = i
// 	nmh[j].index = j
// }

// func (nmh *NodeMinHeap) Push(x any) {
// 	n := len(*nmh)
// 	NodeMinHeapItem := x.(*NodeMinHeapItem)
// 	NodeMinHeapItem.index = n
// 	*nmh = append(*nmh, NodeMinHeapItem)
// }

// func (nmh *NodeMinHeap) Pop() any {
// 	old := *nmh
// 	n := len(old)
// 	NodeMinHeapItem := old[n-1]
// 	old[n-1] = nil             // don't stop the GC from reclaiming the NodeMinHeapItem eventually
// 	NodeMinHeapItem.index = -1 // for safety
// 	*nmh = old[0 : n-1]
// 	return NodeMinHeapItem
// }

/* -----------Strict heap for not letting heap grow max size ------------- */

type BoundedNodeHeap struct {
	items     []*NodeMinHeapItem
	target    *Node
	maxSize   int
	contacted map[string]struct{}
}

func NewBoundedNodeHeap(target *Node, maxSize int) *BoundedNodeHeap {
	h := &BoundedNodeHeap{
		target:    target,
		maxSize:   maxSize,
		contacted: make(map[string]struct{}),
	}
	heap.Init(h)
	return h
}

// Len is part of heap.Interface.
func (h BoundedNodeHeap) Len() int {
	return len(h.items)
}

// Less is part of heap.Interface.
// We want a *max-heap* by distance, so the farthest node is at index 0.
// bytes.Compare(a, b) > 0 means a > b.
func (h BoundedNodeHeap) Less(i, j int) bool {
	return bytes.Compare(h.items[i].distance, h.items[j].distance) == 1
}

// Swap is part of heap.Interface.
func (h BoundedNodeHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

// Push is part of heap.Interface.
func (h *BoundedNodeHeap) Push(x any) {
	n := len(h.items)
	item := x.(*NodeMinHeapItem)
	item.index = n
	h.items = append(h.items, item)
}

// Pop is part of heap.Interface.
func (h *BoundedNodeHeap) Pop() any {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	h.items = old[:n-1]
	return item
}

// check for existence of a node in the heap
func (h *BoundedNodeHeap) contains(n *Node) bool {
	for _, it := range h.items {
		if bytes.Equal(it.node.Node_id, n.Node_id) {
			return true
		}
	}
	return false
}

// AddNode adds a node to the bounded heap if it's not already present
func (h *BoundedNodeHeap) AddNode(n *Node) {
	if h.contains(n) {
		return
	}

	dist := h.target.GetXorDistance(n)

	// If we don't have enough nodes yet, just push.
	if len(h.items) < h.maxSize {
		heap.Push(h, &NodeMinHeapItem{
			node:     *n,
			distance: dist,
		})
		return
	}

	// Heap is full. Root is the *farthest* node (because it's a max-heap).
	worst := h.items[0]

	// If the new node is farther or equal -> ignore it
	if bytes.Compare(dist, worst.distance) >= 0 {
		return
	}

	// New node is closer: remove farthest, then insert this one
	heap.Pop(h) // remove worst
	heap.Push(h, &NodeMinHeapItem{
		node:     *n,
		distance: dist,
	})
}

func (h *BoundedNodeHeap) MarkContacted(n *Node) {
	h.contacted[string(n.Node_id)] = struct{}{}
}

func (h *BoundedNodeHeap) GetUncontacted() []*Node {
	var out []*Node
	for _, it := range h.items {
		if _, ok := h.contacted[string(it.node.Node_id)]; !ok {
			out = append(out, &it.node)
		}
	}
	return out
}

func (h *BoundedNodeHeap) HaveContactedAll() bool {
	return len(h.GetUncontacted()) == 0
}

func (h *BoundedNodeHeap) Closest() []*Node {
	tmp := make([]*NodeMinHeapItem, len(h.items))
	copy(tmp, h.items)

	sort.Slice(tmp, func(i, j int) bool {
		return bytes.Compare(tmp[i].distance, tmp[j].distance) < 0
	})

	out := make([]*Node, 0, len(tmp))
	for _, it := range tmp {
		out = append(out, &it.node)
	}
	return out
}

//// sample to run  --------------------------

// bh := NewBoundedNodeHeap(targetNode, K)

// for _, n := range initialNodes {
//     bh.AddNode(n)
// }

// for !bh.HaveContactedAll() {
//     uncontacted := bh.GetUncontacted()
//     // pick α of them, send RPCs, mark contacted, AddNode(newNodes...)
// }
