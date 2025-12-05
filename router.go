package main

import (
	"fmt"
	"math/big"
	"slices"
	"time"
	
)

type Router struct {
	node Node
	// protocol Protocol
	buckets []*KBucket
}

func NewRouter(node Node) Router {
	router := Router{
		node:    node,
		buckets: nil,
	}
	router.FlushCache()

	fmt.Println("buckets: ", router.buckets)

	return router
}

func (self *Router) FlushCache() {
	lower := big.NewInt(0)
	upper := big.NewInt(1)
	upper.Lsh(upper, NODE_ID_BIT_SIZE)
	all_encompassing_bucket := NewKBucket(lower, upper)
	self.buckets = append(self.buckets, &all_encompassing_bucket)
}

func (self *Router) SplitBucket(index int) {
	first, second := self.buckets[index].Split()
	self.buckets[index] = &first
	self.buckets = slices.Insert(self.buckets, index, &second)
}

func (self *Router) LonelyBuckets() []*KBucket {
	now := time.Now()
	// find buckets which haven't been updated since an hour
	hourago := now.Add(time.Hour * -1)

	lonelyBuckets := make([]*KBucket, 0, 1)

	for _, bucket := range self.buckets {
		if bucket.last_updated.Before(hourago) {
			lonelyBuckets = append(lonelyBuckets, bucket)
		}
	}

	return lonelyBuckets
}

func (self *Router) IsNewNode(n Node) bool {
	index := self.GetBucketFor(n)
	if index == -1 {
		return true
	}
	return self.buckets[index].IsNewNode(n.HexID())
}

func (self *Router) RemoveContact(n Node) {
	index := self.GetBucketFor(n)
	if index == -1 {
		return
	}
	self.buckets[index].RemoveNode(n)
}

func (self *Router) AddContact(n Node) {
	index := self.GetBucketFor(n)
	if index == -1 {
		return
	}
	bucket := self.buckets[index]

	if bucket.AddNode(n) {
		fmt.Println("router: added contact successfully: ", n.HexID())
		return
	}

	// if we are here, the bucket was full and addNode failed
	// split the bucket if it has the router node in its range
	// or if its depth is not congruent to 0, mod BSIZE

	fmt.Println("adding contact did not succeed - bucket full, splitting")
	if bucket.HasInRange(self.node.nodeID) || bucket.Depth()%BSIZE != 0 {
		self.SplitBucket(index)
		self.AddContact(n)
	} else {
		//TODO: ping the head of the bucket list
	}
}

func (self *Router) GetBucketFor(n Node) int {
	for index, bucket := range self.buckets {
		if bucket.HasInRange(n.nodeID) {
			return index
		}
	}

	return -1
}

func (self *Router) FindNeighbors(n Node, alpha int) []*Node {
	heapsize := alpha
	if alpha == -1 || alpha <= 0 {
		heapsize = KSIZE
	}

	nodes := NewBoundedNodeHeap(&n, heapsize)

	traverser := NewTraversal(self, n)

	for {
		neighbor, done := traverser.Next()
		if done {
			break
		}

		if neighbor.nodeID == nil || neighbor.HexID() == n.HexID() {
			continue
		}

		nodes.AddNode(&neighbor)
		if nodes.Len() == heapsize {
			break
		}
	}

	return nodes.Closest()
}

type Traversal struct {
	currentNodes []Node
	leftBuckets  []*KBucket
	rightBuckets []*KBucket
	curr_index   int
	left_index   int
	right_index  int
	isLeft       bool
}

func NewTraversal(router *Router, startNode Node) *Traversal {
	index := router.GetBucketFor(startNode)
	router.buckets[index].RefreshLastUpdated()
	currentNodes := router.buckets[index].GetNodes()
	leftBuckets := router.buckets[:index]
	rightBuckets := router.buckets[index+1:]

	t := &Traversal{
		currentNodes: currentNodes,
		leftBuckets:  leftBuckets,
		rightBuckets: rightBuckets,
		curr_index:   len(currentNodes) - 1,
		left_index:   len(leftBuckets) - 1, // start at last left bucket
		right_index:  0,                    // start at first right bucket
		isLeft:       true,
	}

	return t
}

func (self *Traversal) Next() (Node, bool) {
	if self.curr_index >= 0 {
		res := self.currentNodes[self.curr_index]
		self.curr_index--
		return res, false
	}

	if self.isLeft && self.left_index >= 0 {
		self.currentNodes = self.leftBuckets[self.left_index].GetNodes()
		self.left_index--
		self.curr_index = len(self.currentNodes) - 1
		self.isLeft = false
		return self.Next()
	}

	if self.right_index < len(self.rightBuckets) {
		self.currentNodes = self.rightBuckets[self.right_index].GetNodes()
		self.right_index++
		self.curr_index = len(self.currentNodes) - 1
		self.isLeft = true
		return self.Next()
	}

	return Node{}, true

}
