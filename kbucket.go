package main

import (
	"github.com/matheusoliveira/go-ordered-map/omap"
	//"github.com/amjadjibon/itertools"
	"time"
	//"slices"
	"strings"
	"bytes"
)

type KBucket struct {
	range_lower NodeID
	range_upper NodeID
	nodelist omap.OMap[string, Node]
	last_updated time.Time
	replacement_nodelist omap.OMap[string, Node]
	max_replacment_nodes int
}

func NewKBucket(range_lower NodeID, range_upper NodeID) (KBucket) {

	// make node lists
	_nodelist := omap.New[string, Node]()
	_replacement_nodelist := omap.New[string, Node]()

	return KBucket {
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
		if first.HasInRange(it.Value().node_id) {
			first.nodelist.Put(it.Key(), it.Value())
		} else {
			second.nodelist.Put(it.Key(), it.Value())
		}
		
	}

	for it := self.replacement_nodelist.Iterator(); it.Next(); {
		if first.HasInRange(it.Value().node_id) {
			first.nodelist.Put(it.Key(), it.Value())
		} else {
			second.nodelist.Put(it.Key(), it.Value())
		}
		
	}

	return first, second

}

func (self *KBucket) GetNodes() ([]Node) {
	return omap.IteratorValuesToSlice(self.nodelist.Iterator())
}

func (self *KBucket) AddNode(n Node) (bool) {
	_, found := self.nodelist.Get(n.HexID())
	if found {
		// delete the node and re-add if it exists, to preserve the order of last seen
		self.nodelist.Delete(n.HexID())
		self.nodelist.Put(n.HexID(), n)
	} else if self.Len() < KSIZE {
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
			newest_seen := omap.IteratorKeysToSlice(self.replacement_nodelist.Iterator())[self.replacement_nodelist.Len() - 1]
			// add to node list
			self.nodelist.Put(newest_seen, n)
		}
	}
}

func (self *KBucket) GetNode(node_id string) (Node) {
	node, _ := self.nodelist.Get(node_id)
	return node
}

func (self *KBucket) IsNewNode(node_id string) bool {
	_, found := self.nodelist.Get(node_id)
	return !found
}

func (self *KBucket) HasInRange(node_id NodeID) bool {
	return bytes.Compare(node_id, self.range_lower) > -1 && bytes.Compare(self.range_upper, node_id) > -1
}

// TODO double check this is actually the oldest seen
func (self *KBucket) Head() (Node) {
	head := omap.IteratorValuesToSlice(self.nodelist.Iterator())[0]
	return head
}

func (self *KBucket) Len() int {
	return self.nodelist.Len()
}

func (self *KBucket) Depth() int {
	node_ids := omap.IteratorKeysToSlice(self.nodelist.Iterator())
	sp := SharedPrefix(node_ids)
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