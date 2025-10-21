package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/go-faster/xor"
)

const NODE_ID_BUFFER_SIZE int = 20 // 20 bytes in 160-bit node ID
type NodeID [NODE_ID_BUFFER_SIZE]byte

func main() {
	x := make([]byte, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	y := make([]byte, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)

	// Initializing the elements
	for i := 0; i < 10; i++ {
		x[i] = byte(i + 1)
		y[i] = byte(i + 1)
	}

	res := xor_distance(x, y)
	fmt.Println(res)

	id1, _ := NewNodeIDFromUDP("192.0.2.10", 4001)
	id2, _ := NewNodeIDFromUDP("2001:db8::1", 4001)
	fmt.Println("ID1:", id1)
	fmt.Println("ID2:", id2)
}

// assuming network byte order (big endian)
func xor_distance(x, y []byte) []byte {
	fmt.Println(x)
	fmt.Println(y)

	res := make([]byte, 20, 20)
	xor.Bytes(res, x, y)
	return res
}

// String returns the hexadecimal representation of the NodeID.
func (n NodeID) String() string { return hex.EncodeToString(n[:]) }

// Using Ip address and UDP port to generate NodeID
func NewNodeIDFromUDP(ipStr string, port int, extra ...[]byte) (NodeID, error) {
	var id NodeID

	// Parse the IP address and Canonicalize it
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return id, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Cannonicalize to 16 bytes. For IPv4, map to v4-in-v6
	ip16 := ip.To16()
	if ip16 == nil {
		return id, fmt.Errorf("unable to convert IP to 16-byte representation: %s", ipStr)
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
	copy(id[:], sum[:NODE_ID_BUFFER_SIZE]) // take first 20 bytes
	return id, nil

}
