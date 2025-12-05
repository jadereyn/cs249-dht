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
	Router    *Router
	// Replace with actual struct later
	Store map[string][]byte
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

	router := NewRouter(selfNode)

	return &LocalNode{
		Self:      selfNode,
		Transport: transport,
		Router:    &router,
		Store:     make(map[string][]byte),
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

	remoteNode, err := NodeFromRPC(msg)
	if err == nil {
		ln.Router.AddContact(*remoteNode)
	}

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
		if err := ln.sendDirectRPC(pong, from); err != nil {
			fmt.Printf("Error sending Pong RPC: %v\n", err)
		}

	case RPCPong:
		fmt.Printf("LocalNode %s received Pong from %s:%d (id=%s)\n",
			ln.Self.HexID(), msg.FromIP, msg.FromPort, msg.FromID)

	case RPCFindNode:
		fmt.Printf("LocalNode %s got FIND_NODE (not implemented yet)\n",
			ln.Self.HexID())
		// TODO: use routing table to reply with closest nodes
		ln.handleFindNodeRPC(msg, from)

	case RPCStore:
		fmt.Printf("LocalNode %s got STORE key=%q\n",
			ln.Self.HexID(), msg.Key)
		// TODO: store key/value
		if msg.Key == "" {
			fmt.Println("STORE with empty key, ignoring")
			return
		}
		ln.StoreLocal(msg.Key, msg.Value)

		// optional: send a simple ACK (not required by spec, but handy)
		// TODO: what is this doing and what do we need it for?
		ack := &RPCMessage{
			Type:     RPCStore, // or define RPCStoreAck if you want
			FromID:   ln.Self.HexID(),
			FromIP:   ln.Self.ipAddr,
			FromPort: ln.Self.port,
			Key:      msg.Key,
		}
		if err := ln.sendDirectRPC(ack, from); err != nil {
			fmt.Printf("Error sending STORE ack: %v\n", err)
		}

	case RPCFindValue:
		fmt.Printf("LocalNode %s got FIND_VALUE for key=%q\n",
			ln.Self.HexID(), msg.Key)

		ln.handleFindValueRPC(msg, from)

	default:
		fmt.Printf("LocalNode %s got unknown RPC type %v\n",
			ln.Self.HexID(), msg.Type)
	}
}

func (ln *LocalNode) sendDirectRPC(msg *RPCMessage, to *net.UDPAddr) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal rpc: %w", err)
	}
	_, err = ln.Transport.conn.WriteToUDP(payload, to)
	if err != nil {
		return fmt.Errorf("write to udp: %w", err)
	}
	return nil
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

	// Build a Node for the bootstrap and add to routing table.
	bootstrapNode, err := NodeFromRPC(resp)
	if err != nil {
		fmt.Printf("LocalNode %s: error converting Ping response to Node: %v\n",
			ln.Self.HexID(), err)
		return
	}
	ln.Router.AddContact(*bootstrapNode)

}

// Convert our internal Node to wire format
func NodeToRPC(n *Node) RPCNodeInfo {
	return RPCNodeInfo{
		ID:   n.HexID(),
		IP:   n.ipAddr,
		Port: n.port,
	}
}

// Convert the RPC sender info back into a Node
func NodeFromRPC(msg *RPCMessage) (*Node, error) {
	if msg.FromID == "" {
		return nil, fmt.Errorf("RPCMessage.FromID is empty")
	}
	id := new(big.Int)
	_, ok := id.SetString(msg.FromID, 16)
	if !ok {
		return nil, fmt.Errorf("invalid FromID hex: %s", msg.FromID)
	}
	return &Node{
		ipAddr: msg.FromIP,
		port:   msg.FromPort,
		nodeID: id,
	}, nil
}

// Handle FindNode RPC by looking up closest nodes and replying.
func (ln *LocalNode) handleFindNodeRPC(msg *RPCMessage, from *net.UDPAddr) {
	if msg.TargetID == "" {
		fmt.Println("FindNode RPC with empty TargetID")
		return
	}

	// Parse target ID from hex
	targetID := new(big.Int)
	if _, ok := targetID.SetString(msg.TargetID, 16); !ok {
		fmt.Printf("Invalid TargetID hex in FindNode: %s\n", msg.TargetID)
		return
	}

	// Make a dummy Node with this ID just to use Router.FindNeighbors
	targetNode := Node{
		ipAddr: "", // not used for distance
		port:   0,  // not used for distance
		nodeID: targetID,
	}

	// Use your routing table + kbuckets to find nearest neighbors
	neighbors := ln.Router.FindNeighbors(targetNode, -1) // -1 => use KSIZE internally

	// Convert to RPCNodeInfo for the wire
	nodeInfos := make([]RPCNodeInfo, 0, len(neighbors))
	for _, n := range neighbors {
		nodeInfos = append(nodeInfos, NodeToRPC(n))
	}

	resp := &RPCMessage{
		Type:     RPCFindNode, // or RPCFindNodeResp if you add a separate type
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
		TargetID: msg.TargetID,
		Nodes:    nodeInfos,
	}

	if err := ln.sendDirectRPC(resp, from); err != nil {
		fmt.Printf("Error sending FindNode response: %v\n", err)
	}
}

// FindNodeOnce sends a single FindNode RPC to the given ip/port and returns the neighbors.
func (ln *LocalNode) FindNodeOnce(targetID *big.Int, ip string, port int) ([]Node, error) {
	msg := &RPCMessage{
		Type:     RPCFindNode,
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
		TargetID: NodeIDToHex(targetID),
	}

	resp, err := ln.Transport.SendRPC(ip, port, msg, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("SendRPC FindNode: %w", err)
	}

	neighbors := make([]Node, 0, len(resp.Nodes))
	for _, info := range resp.Nodes {
		id := new(big.Int)
		if _, ok := id.SetString(info.ID, 16); !ok {
			continue
		}

		n := Node{
			ipAddr: info.IP,
			port:   info.Port,
			nodeID: id,
		}

		// Learn about this contact too
		ln.Router.AddContact(n)

		neighbors = append(neighbors, n)
	}

	return neighbors, nil
}

// LookupNodes performs a Kademlia-style iterative lookup for nodes
// close to targetID, and returns up to KSIZE closest nodes it finds.
func (ln *LocalNode) LookupNodes(targetID *big.Int) ([]Node, error) {
	// 1. Start from our own routing table
	targetNode := Node{
		ipAddr: "",
		port:   0,
		nodeID: targetID,
	}

	initial := ln.Router.FindNeighbors(targetNode, KSIZE)
	if len(initial) == 0 {
		return nil, fmt.Errorf("no known nodes in routing table")
	}

	// 2. Create a bounded heap keyed by distance to target
	heap := NewBoundedNodeHeap(&targetNode, KSIZE)
	for _, n := range initial {
		if n == nil || n.nodeID == nil {
			continue
		}
		heap.AddNode(n)
	}

	// Track which nodes we've already queried
	tried := make(map[string]bool)

	for {
		// 3. Get uncontacted nodes, closest first
		uncontacted := heap.GetUncontacted()
		if len(uncontacted) == 0 {
			break
		}

		// Take up to ALPHA at a time
		batch := uncontacted
		if len(batch) > ALPHA {
			batch = batch[:ALPHA]
		}

		progress := false

		for _, n := range batch {
			if n == nil || n.nodeID == nil {
				continue
			}
			idHex := n.HexID()
			if tried[idHex] {
				continue
			}
			tried[idHex] = true
			heap.MarkContacted(n)

			// 4. Ask this node for neighbors of targetID
			newNodes, err := ln.FindNodeOnce(targetID, n.ipAddr, n.port)
			if err != nil {
				// errors are common (timeouts, offline nodes), just skip
				continue
			}

			// 5. Merge newly discovered nodes
			for _, nn := range newNodes {
				// Make sure we don't freak out if nodeID is nil
				if nn.nodeID == nil {
					continue
				}
				heap.AddNode(&nn)
				ln.Router.AddContact(nn)
			}

			if len(newNodes) > 0 {
				progress = true
			}
		}

		// 6. If none of the batch gave us new nodes, we converged
		if !progress {
			break
		}
	}

	// 7. Return the K closest nodes from heap
	closestPtrs := heap.Closest() // []*Node
	out := make([]Node, 0, len(closestPtrs))
	for _, p := range closestPtrs {
		if p != nil && p.nodeID != nil {
			out = append(out, *p)
		}
	}
	return out, nil
}

// TODO : we might not need them if we use our own struct for storage
// StoreLocal stores a key-value pair in the local node's storage.
func (ln *LocalNode) StoreLocal(key string, value []byte) {
	ln.Store[key] = value
}

// GetLocal retrieves a value by key from the local node's storage.
func (ln *LocalNode) GetLocal(key string) ([]byte, bool) {
	v, ok := ln.Store[key]
	return v, ok
}

func (ln *LocalNode) handleFindValueRPC(msg *RPCMessage, from *net.UDPAddr) {
	if msg.Key == "" {
		fmt.Println("FindValue RPC with empty key")
		return
	}

	// 1) If we *have* the value locally, return it directly.
	if val, ok := ln.GetLocal(msg.Key); ok {
		resp := &RPCMessage{
			Type:     RPCFindValue,
			FromID:   ln.Self.HexID(),
			FromIP:   ln.Self.ipAddr,
			FromPort: ln.Self.port,
			Key:      msg.Key,
			Value:    val,
			// Nodes can be empty when value is returned
		}
		if err := ln.sendDirectRPC(resp, from); err != nil {
			fmt.Printf("Error sending FindValue value response: %v\n", err)
		}
		return
	}

	// 2) Otherwise, behave like FIND_NODE on the key’s ID.

	// Derive an ID from the key (SHA-256 just like Node IDs)
	keyHash := sha256.Sum256([]byte(msg.Key))
	keyID := new(big.Int).SetBytes(keyHash[:])

	targetNode := Node{
		ipAddr: "",
		port:   0,
		nodeID: keyID,
	}

	neighbors := ln.Router.FindNeighbors(targetNode, -1)

	nodeInfos := make([]RPCNodeInfo, 0, len(neighbors))
	for _, n := range neighbors {
		nodeInfos = append(nodeInfos, NodeToRPC(n))
	}

	resp := &RPCMessage{
		Type:     RPCFindValue, // same type; distinguish by Value vs Nodes
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
		Key:      msg.Key,
		Nodes:    nodeInfos,
	}

	if err := ln.sendDirectRPC(resp, from); err != nil {
		fmt.Printf("Error sending FindValue nodes response: %v\n", err)
	}
}

func (ln *LocalNode) StoreValue(key string, value []byte) error {
	// Hash the key into an ID in the same space as node IDs
	keyHash := sha256.Sum256([]byte(key))
	keyID := new(big.Int).SetBytes(keyHash[:])

	// Ask our own router for KSIZE closest nodes
	targetNode := Node{nodeID: keyID}
	neighbors := ln.Router.FindNeighbors(targetNode, KSIZE)
	if len(neighbors) == 0 {
		return fmt.Errorf("StoreValue: no known nodes to store to")
	}

	msg := &RPCMessage{
		Type:     RPCStore,
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
		Key:      key,
		Value:    value,
	}

	// Fire STORE RPC to each neighbor (we can ignore acks for now)
	for _, n := range neighbors {
		if n == nil || n.nodeID == nil {
			continue
		}
		_, err := ln.Transport.SendRPC(n.ipAddr, n.port, msg, 3*time.Second)
		if err != nil {
			// not fatal; some nodes may be down
			fmt.Printf("StoreValue: error storing to %s:%d: %v\n",
				n.ipAddr, n.port, err)
		}
	}

	// Optionally also store locally
	ln.StoreLocal(key, value)

	return nil
}

func (ln *LocalNode) FindValueOnce(key string, ip string, port int) (value []byte, nodes []Node, err error) {
	msg := &RPCMessage{
		Type:     RPCFindValue,
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
		Key:      key,
	}

	resp, err := ln.Transport.SendRPC(ip, port, msg, 5*time.Second)
	if err != nil {
		return nil, nil, fmt.Errorf("SendRPC FindValue: %w", err)
	}

	// If the value field is non-empty, we’re done.
	if len(resp.Value) > 0 {
		return resp.Value, nil, nil
	}

	// Otherwise, convert resp.Nodes to []Node (just like FindNodeOnce)
	out := make([]Node, 0, len(resp.Nodes))
	for _, info := range resp.Nodes {
		id := new(big.Int)
		if _, ok := id.SetString(info.ID, 16); !ok {
			continue
		}
		n := Node{
			ipAddr: info.IP,
			port:   info.Port,
			nodeID: id,
		}
		ln.Router.AddContact(n)
		out = append(out, n)
	}

	return nil, out, nil
}
