package main

import (
	"slices"
	"math/big"
	"time"
)

type Router struct {
	node Node
	// protocol Protocol
	buckets []KBucket
}

func NewRouter(node Node) (Router) {

	router := Router {
		node,
		nil,
	}
	router.FlushCache()

	return router
}

func (self *Router) FlushCache() {
	lower := big.NewInt(0)
	upper := big.NewInt(1)
	upper.Lsh(upper, NODE_ID_BIT_SIZE)
	all_encompassing_bucket := NewKBucket(lower, upper)
	self.buckets = make([]KBucket, 1, 1)
	self.buckets = append(self.buckets, all_encompassing_bucket)
}

func (self *Router) SplitBucket(index int) {
	first, second := self.buckets[index].Split()
	self.buckets[index] = first
	slices.Insert(self.buckets, index, second)
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
	return self.buckets[index].IsNewNode(n.HexID())
}

func (self *Router) RemoveContact(n Node) {
	index := self.GetBucketFor(n)
	self.buckets[index].RemoveNode(n)
}

func (self *Router) AddContact(n Node) {
	index := self.GetBucketFor(n)
	bucket := self.buckets[index]

	if bucket.AddNode(n) {
		return
	}

	// if we are here, the bucket was full and addNode failed
	// split the bucket if it has the router node in its range
	// or if its depth is not congruent to 0, mod BSIZE

	if bucket.HasInRange(self.node.nodeID) || bucket.Depth() % BSIZE != 0 {
		self.SplitBucket(index)
		self.AddContact(n)
	} else {
		//TODO: ping the head of the bucket list
	}
}

func (self *Router) GetBucketFor(n Node) int {
	for index, bucket := range self.buckets {
		if n.nodeID.Cmp(bucket.range_upper) == -1 {
			return index
		}
	}

	return -1
}

func (self *Router) FindNeighbors(n Node, alpha int) []*Node {
	
	heapsize := alpha
	if alpha == -1 {
		heapsize = KSIZE
	}

	nodes := NewBoundedNodeHeap(&n, heapsize)

	traverser := NewTraversal(self, n)

	neighbor, isComplete := traverser.Next()

	for !isComplete {
		
		if neighbor.HexID() != n.HexID() {
			nodes.Push(neighbor)
		}

		if nodes.Len() == heapsize {
			break
		}

		neighbor, isComplete = traverser.Next()
	}

	return nodes.Closest()
}


type Traversal struct {
	index int
	currentNodes []Node
	leftBuckets []KBucket
	rightBuckets []KBucket
	curr_index int
	left_index int
	right_index int
	isLeft bool
}

func NewTraversal(router *Router, startNode Node) Traversal {
	index := router.GetBucketFor(startNode)
	router.buckets[index].RefreshLastUpdated()
	currentNodes := router.buckets[index].GetNodes()
	leftBuckets := router.buckets[:index]
	rightBuckets := router.buckets[index+1:]

	t := Traversal {
		index,
		currentNodes,
		leftBuckets,
		rightBuckets,
		len(currentNodes) - 1,
		len(leftBuckets) - 1, // start at last left bucket
		0, // start at first right bucket
		true,
	}

	return t
}

func (self *Traversal) Next() (Node, bool) {
	if self.curr_index >= 0 {
		res := self.currentNodes[self.curr_index]
		self.curr_index -= 1
		return res, false
	}

	if self.isLeft && self.left_index >= 0 {
		self.currentNodes = self.leftBuckets[self.left_index].GetNodes()
		self.isLeft = false
		return self.Next()
	}

	if self.right_index < len(self.rightBuckets) {
		self.currentNodes = self.leftBuckets[self.left_index].GetNodes()
		self.isLeft = true
		return self.Next()
	}

	return Node{"nil", 0, nil}, true
}