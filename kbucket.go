package main

import (
	"github.com/matheusoliveira/go-ordered-map/omap"
	//"github.com/amjadjibon/itertools"
	"time"
	//"slices"
	"strings"
	//"bytes"
	"math/big"
	"fmt"
)

type KBucket struct {
	range_lower          *big.Int
	range_upper          *big.Int
	nodelist             omap.OMap[string, Node]
	last_updated         time.Time
	replacement_nodelist omap.OMap[string, Node]
	max_replacment_nodes int
}

func NewKBucket(range_lower *big.Int, range_upper *big.Int) KBucket {

	// make node lists
	_nodelist := omap.New[string, Node]()
	_replacement_nodelist := omap.New[string, Node]()

	return KBucket{
		range_lower,
		range_upper,
		_nodelist,
		time.Now(),
		_replacement_nodelist,
		KSIZE * REPLACEMENT_FACTOR,
	}
}

func (self *KBucket) RefreshLastUpdated() {
	self.last_updated = time.Now()
}

func (self *KBucket) Split() (KBucket, KBucket) {
	midp, mplusone := FindMidpoint(self.range_lower, self.range_upper)
	first := NewKBucket(self.range_lower, midp)
	second := NewKBucket(mplusone, self.range_upper)

	// transfer nodes by id here to each bucket
	for it := self.nodelist.Iterator(); it.Next(); {
		if first.HasInRange(it.Value().nodeID) {
			first.nodelist.Put(it.Key(), it.Value())
		} else {
			second.nodelist.Put(it.Key(), it.Value())
		}

	}

	for it := self.replacement_nodelist.Iterator(); it.Next(); {
		if first.HasInRange(it.Value().nodeID) {
			first.nodelist.Put(it.Key(), it.Value())
		} else {
			second.nodelist.Put(it.Key(), it.Value())
		}

	}

	return first, second

}

func (self *KBucket) GetNodes() []Node {
	return omap.IteratorValuesToSlice(self.nodelist.Iterator())
}

func (self *KBucket) GetReplacementNodes() []Node {
	return omap.IteratorValuesToSlice(self.replacement_nodelist.Iterator())
}

func (self *KBucket) AddNode(n Node) bool {
	_, found := self.nodelist.Get(n.HexID())
	if found {
		// delete the node and re-add if it exists, to preserve the order of last seen
		self.nodelist.Delete(n.HexID())
		self.nodelist.Put(n.HexID(), n)
	} else if self.Len() < KSIZE {
		//fmt.Println("bucket not yet full, ", n.HexID())
		self.nodelist.Put(n.HexID(), n)
	} else {
		_, found = self.replacement_nodelist.Get(n.HexID())
		if found {
			self.replacement_nodelist.Delete(n.HexID())
		}
		self.replacement_nodelist.Put(n.HexID(), n)

		for self.replacement_nodelist.Len() > self.max_replacment_nodes {
			oldest_seen := omap.IteratorKeysToSlice(self.replacement_nodelist.Iterator())[0]
			self.replacement_nodelist.Delete(oldest_seen)
		}

		fmt.Println("bucket full, should return false, ", n.HexID())

		return false
	}

	return true
}

func (self *KBucket) RemoveNode(n Node) {
	_, found := self.replacement_nodelist.Get(n.HexID())
	if found {
		self.replacement_nodelist.Delete(n.HexID())
	}

	_, found = self.nodelist.Get(n.HexID())
	if found {
		self.nodelist.Delete(n.HexID())

		if self.replacement_nodelist.Len() > 0 {
			// get newest seen (last added)
			newest_seen_id := omap.IteratorKeysToSlice(self.replacement_nodelist.Iterator())[self.replacement_nodelist.Len()-1]
			// add to node list
			newest_seen_node, _ := self.replacement_nodelist.Get(newest_seen_id)
			//fmt.Println("newest_seen_id: ", newest_seen_id)
			//fmt.Println("node gotten: ", newest_seen_node.HexID())

			self.nodelist.Put(newest_seen_id, newest_seen_node)

			// remove newest seen replacement node from replacement node list
			self.replacement_nodelist.Delete(newest_seen_id)
		}
	}
}

func (self *KBucket) GetNode(nodeID string) Node {
	node, _ := self.nodelist.Get(nodeID)
	return node
}

func (self *KBucket) IsNewNode(nodeID string) bool {
	_, found := self.nodelist.Get(nodeID)
	return !found
}

func (self *KBucket) HasInRange(nodeID *big.Int) bool {
	//return bytes.Compare(nodeID, self.range_lower) > -1 && bytes.Compare(self.range_upper, nodeID) > -1
	return nodeID.Cmp(self.range_lower) > -1 && self.range_upper.Cmp(nodeID) > -1
}

// TODO double check this is actually the oldest seen
func (self *KBucket) Head() Node {
	head := omap.IteratorValuesToSlice(self.nodelist.Iterator())[0]
	return head
}

func (self *KBucket) Len() int {
	return self.nodelist.Len()
}

func (self *KBucket) Depth() int {
	nodeIDs := omap.IteratorKeysToSlice(self.nodelist.Iterator())
	sp := SharedPrefix(nodeIDs)
	return len(sp)
}

func SharedPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]

	for _, str := range strs {
		for !strings.HasPrefix(str, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix
}
