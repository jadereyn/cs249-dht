package main

import (
	"math/big"
	"time"
)

type Router struct {
	node Node
	// protocol Protocol
	buckets []KBucket
}

func NewRouter(node Node) Router {
	router := Router{
		node:    node,
		buckets: nil,
	}
	router.FlushCache()

	return router
}

func (self *Router) FlushCache() {
	lower := big.NewInt(0)
	upper := big.NewInt(1)
	upper.Lsh(upper, NODE_ID_BIT_SIZE)

	all := NewKBucket(lower, upper)
	self.buckets = make([]KBucket, 0, 1)
	self.buckets = append(self.buckets, all)
}

func (self *Router) SplitBucket(index int) {
	first, second := self.buckets[index].Split()
	self.buckets[index] = first

	oldLen := len(self.buckets)
	self.buckets = append(self.buckets, KBucket{})
	copy(self.buckets[index+2:], self.buckets[index+1:oldLen])
	self.buckets[index+1] = second
}

func (self *Router) LonelyBuckets() []KBucket {
	now := time.Now()
	// find buckets which haven't been updated since an hour
	hourago := now.Add(time.Hour * -1)

	lonelyBuckets := make([]KBucket, 0, 1)

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
	bucket := &self.buckets[index]

	if bucket.AddNode(n) {
		return
	}

	// if we are here, the bucket was full and addNode failed
	// split the bucket if it has the router node in its range
	// or if its depth is not congruent to 0, mod BSIZE

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
	leftBuckets  []KBucket
	rightBuckets []KBucket
	currIndex    int
	leftIndex    int
	rightIndex   int
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
		currIndex:    len(currentNodes) - 1,
		leftIndex:    len(leftBuckets) - 1, // start at last left bucket
		rightIndex:   0,                    // start at first right bucket
		isLeft:       true,
	}

	return t
}

func (self *Traversal) Next() (Node, bool) {
	for {
		if self.currIndex >= 0 {
			res := self.currentNodes[self.currIndex]
			self.currIndex--
			return res, false
		}

		if self.isLeft && self.leftIndex >= 0 {
			self.currentNodes = self.leftBuckets[self.leftIndex].GetNodes()
			self.currIndex = len(self.currentNodes) - 1
			self.leftIndex--
			self.isLeft = false
			continue
		}

		if self.rightIndex < len(self.rightBuckets) {
			self.currentNodes = self.rightBuckets[self.rightIndex].GetNodes()
			self.currIndex = len(self.currentNodes) - 1
			self.rightIndex++
			self.isLeft = true
			continue
		}

		return Node{}, true
	}
}
