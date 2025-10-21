package main

type KBucket struct {
	nodes []Node
}

type Node struct {
	
	ip_addr string
	hash_ID []byte
	k_buckets []KBucket

}

func (self *Node) GenerateID() {
	// - get IP address of node
	// - hash IP address with some hashing algorithm
	// - truncate to first 160 bits
}


func (self *Node) GetXorDistance(n *Node) {
	// - return xor distance from self to n
}


func (self *Node) Ping(n *Node) {

}

// func (n *Node) Store(n *node, k *key, v *value) {
	
// }

func (self *Node) FindNode(hashID []byte) {

}

func (self *Node) FindValue() {

}