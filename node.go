package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"errors"

	//"log/slog"

	"github.com/go-faster/xor"
)

const NODE_ID_BUFFER_SIZE int = 32 // 20 bytes in 160-bit node ID, but we are using sha-256 so change to 32 bytes
type NodeID = []byte

type Node struct {
	
	ip_addr string
	port int
	node_id NodeID

}

// Using Ip address and UDP port to generate new Node
func NewNodeFromIPAndPort(ipStr string, port int, extra ...[]byte) (Node, error) {

	// Parse the IP address and Canonicalize it
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		//return id, fmt.Errorf("invalid IP address: %s", ipStr)
		return Node{}, errors.New(fmt.Sprintf("invalid IP address: %s", ipStr))
	}

	// Canonicalize to 16 bytes. For IPv4, map to v4-in-v6
	ip16 := ip.To16()
	if ip16 == nil {
		//return id, fmt.Errorf("unable to convert IP to 16-byte representation: %s", ipStr)
		return Node{}, errors.New(fmt.Sprintf("unable to convert IP to 16-byte representation: %s", ipStr))
	}

	// Address family marker : 4 or 6 (helps avoid weird collisions)
	var af byte = 6
	if ip.To4() != nil {
		af = 4
	}

	// Build canonical input buffer
	// version byte lets us change the schema later without breaking determinism
	buf := make([]byte, 0, 1+1+16+2+64) // max extra size is 64 bytes
	buf = append(buf, 1)                // version byte
	buf = append(buf, af)               // address family byte (either ip4 or ip6``)
	buf = append(buf, ip16...)          // 16 bytes for IP
	var p [2]byte
	binary.BigEndian.PutUint16(p[:], uint16(port))
	buf = append(buf, p[:]...) // 2 bytes for port

	// Optional extras (salt/nonce/pk hash)
	for _, extraBytes := range extra {
		buf = append(buf, extraBytes...)
	}

	// Hash the buffer to get NodeID
	sum := sha256.Sum256(buf)
	id := make(NodeID, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	copy(id[:], sum[:NODE_ID_BUFFER_SIZE]) // originally here to take first 20 bytes (for 160 bit IDs) but since upgrading to sha-256, using full result

	return Node {ipStr, port, id}, nil
}

// Return xor distance from self to n
func (self *Node) GetXorDistance(n *Node) NodeID {
	
	res := make(NodeID, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	xor.Bytes(res, self.node_id, n.node_id)
	return res
}

// Return the hexadecimal representation of a Node's id.
func (self *Node) HexID() string { 

	return hex.EncodeToString(self.node_id[:]) 
}

// Return the hexadecimal representation of a NodeID value.
func NodeIDToHex(id NodeID) string { 

	return hex.EncodeToString(id[:]) 
}