package main

import (
	"slices"
)

type Router struct {
	node Node
	// protocol Protocol
	buckets []KBucket
}

func NewRouter(node Node) (Router) {


}

func (self *Router) FlushCache() {
	//all_encompassing_bucket = 
}

func (self *Router) SplitBucket(index int) {
	first, second = self.buckets[index].Split()
	self.buckets[index] = first
	slices.Insert(self.buckets, index, second)
}

