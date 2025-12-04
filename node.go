package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	//"log/slog"

	//"github.com/go-faster/xor"
	"math/big"
)

// TODO: change NodeID to fixed size byte array like [32]byte for sha-256
// type NodeID [NODE_ID_BUFFER_SIZE]byte

type Node struct {
	ipAddr string
	port   int
	nodeID *big.Int
}

// LocalNode represents a running Kademlia node on this machine.
// It owns:
//   - a Node identity (ip/port/nodeID)
//   - a UDPTransport (socket)
//   - later: routing table, storage, etc.
type LocalNode struct {
	Self      Node
	Transport *UDPTransport
	// Routing *RoutingTable // hook your k-buckets here later
}

// NewLocalNode builds a Node identity from (ip,port) and binds UDPTransport.
func NewLocalNode(ip string, port int) (*LocalNode, error) {
	// Derive the node ID from ip+port using your existing function
	selfNode, err := NewNodeFromIPAndport(ip, port)
	if err != nil {
		return nil, err
	}

	// Create the UDP transport on the same ip/port
	transport, err := NewUDPTransport(ip, port)
	if err != nil {
		return nil, err
	}

	return &LocalNode{
		Self:      selfNode,
		Transport: transport,
	}, nil
}

// Using Ip address and UDP port to generate new Node
func NewNodeFromIPAndport(ipStr string, port int, extra ...[]byte) (Node, error) {

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
	// id := make(NodeID, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	// copy(id[:], sum[:NODE_ID_BUFFER_SIZE]) // originally here to take first 20 bytes (for 160 bit IDs) but since upgrading to sha-256, using full result
	id := new(big.Int).SetBytes(sum[:])

	return Node{ipStr, port, id}, nil
}

// Return xor distance from self to n
func (self *Node) GetXorDistance(n *Node) *big.Int {

	res := new(big.Int)

	//buf := make(NodeID, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	//xor.Bytes(buf, self.nodeID.FillBytes, n.nodeID)
	res.Xor(self.nodeID, n.nodeID)
	return res
}

// Return the hexadecimal representation of a Node's id.
func (self *Node) HexID() string {

	return NodeIDToHex(self.nodeID)
}

// Return the hexadecimal representation of a NodeID value.
func NodeIDToHex(id *big.Int) string {

	res := make([]byte, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
	res = id.FillBytes(res)
	return hex.EncodeToString(res)
}

func FindMidpoint(n1 *big.Int, n2 *big.Int) (*big.Int, *big.Int) {

	res := new(big.Int)
	resp1 := new(big.Int)

	res.Add(n1, n2)

	// divide by 2
	res.Rsh(res, 2)

	// get mindpoint plus 1
	resp1.SetInt64(1)
	resp1.Add(res, resp1)

	return res, resp1
}

// HandleRPC is called whenever an RPCMessage is received over UDP.
func (ln *LocalNode) HandleRPC(msg *RPCMessage, from *net.UDPAddr) {
	fmt.Printf("LocalNode %s handling RPC type=%v from %v\n",
		ln.Self.HexID(), msg.Type, from)

	// Later: update routing table with msg.FromID/msg.FromIP/msg.FromPort

	switch msg.Type {
	case RPCPing:
		// Reply with Pong
		pong := &RPCMessage{
			Type:     RPCPong,
			FromID:   ln.Self.HexID(),
			FromIP:   ln.Self.ipAddr,
			FromPort: ln.Self.port,
		}
		ln.sendDirectRPC(pong, from)

	case RPCPong:
		fmt.Printf("LocalNode %s received Pong from %s:%d (id=%s)\n",
			ln.Self.HexID(), msg.FromIP, msg.FromPort, msg.FromID)

	case RPCFindNode:
		fmt.Printf("LocalNode %s got FIND_NODE (not implemented yet)\n",
			ln.Self.HexID())
		// TODO: use routing table to reply with closest nodes

	case RPCStore:
		fmt.Printf("LocalNode %s got STORE key=%q\n",
			ln.Self.HexID(), msg.Key)
		// TODO: store key/value

	case RPCFindValue:
		fmt.Printf("LocalNode %s got FIND_VALUE for key=%q\n",
			ln.Self.HexID(), msg.Key)
		// TODO: return value or closest nodes

	default:
		fmt.Printf("LocalNode %s got unknown RPC type %v\n",
			ln.Self.HexID(), msg.Type)
	}
}

// sendDirectRPC sends an RPCMessage back to a known UDP address
// using the same UDP socket.
func (ln *LocalNode) sendDirectRPC(msg *RPCMessage, to *net.UDPAddr) {
	payload, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("Error marshaling RPC: %v\n", err)
		return
	}
	if _, err := ln.Transport.conn.WriteToUDP(payload, to); err != nil {
		fmt.Printf("Error sending RPC: %v\n", err)
	}
}

// Run starts the main listening loop for this node (blocks forever).
func (ln *LocalNode) Run() {
	ln.Transport.ListenRPC(func(msg *RPCMessage, from *net.UDPAddr) {
		ln.HandleRPC(msg, from)
	})
}

// PingBootstrap sends a Ping RPC to a bootstrap node and waits for response.
func (ln *LocalNode) PingBootstrap(bootstrapIP string, bootstrapPort int) {
	ping := &RPCMessage{
		Type:     RPCPing,
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
	}

	resp, err := ln.Transport.SendRPC(bootstrapIP, bootstrapPort, ping, 5*time.Second)
	if err != nil {
		fmt.Printf("LocalNode %s error pinging bootstrap: %v\n",
			ln.Self.HexID(), err)
		return
	}

	fmt.Printf("LocalNode %s got Ping response: %+v\n",
		ln.Self.HexID(), resp)
}
