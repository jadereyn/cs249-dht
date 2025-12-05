package main

import (
	"net"
	"fmt"
	"encoding/json"
	"time"
)

// Server represents a running Kademlia node on this machine.
// It owns:
//   - a Node identity (ip/port/nodeID)
//   - a UDPTransport (socket)
//   - later: routing table, storage, etc.
type Server struct {
	Self      Node
	Transport *UDPTransport
	// Routing *RoutingTable // hook your k-buckets here later
}

// NewServer builds a Node identity from (ip,port) and binds UDPTransport.
func NewServer(ip string, port int) (*Server, error) {
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

	return &Server{
		Self:      selfNode,
		Transport: transport,
	}, nil
}

// HandleRPC is called whenever an RPCMessage is received over UDP.
func (ln *Server) HandleRPC(msg *RPCMessage, from *net.UDPAddr) {
	fmt.Printf("Server %s handling RPC type=%v from %v\n",
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
		fmt.Printf("Server %s received Pong from %s:%d (id=%s)\n",
			ln.Self.HexID(), msg.FromIP, msg.FromPort, msg.FromID)
		// Add node to routing table

	case RPCFindNode:
		fmt.Printf("Server %s got FIND_NODE (not implemented yet)\n",
			ln.Self.HexID())
		// TODO: use routing table to reply with closest nodes

	case RPCStore:
		fmt.Printf("Server %s got STORE key=%q\n",
			ln.Self.HexID(), msg.Key)
		// TODO: store key/value

	case RPCFindValue:
		fmt.Printf("Server %s got FIND_VALUE for key=%q\n",
			ln.Self.HexID(), msg.Key)
		// TODO: return value or closest nodes

	default:
		fmt.Printf("Server %s got unknown RPC type %v\n",
			ln.Self.HexID(), msg.Type)
	}
}

// sendDirectRPC sends an RPCMessage back to a known UDP address
// using the same UDP socket.
func (ln *Server) sendDirectRPC(msg *RPCMessage, to *net.UDPAddr) {
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
func (ln *Server) Run() {
	ln.Transport.ListenRPC(func(msg *RPCMessage, from *net.UDPAddr) {
		go ln.HandleRPC(msg, from)
	})
}

// PingBootstrap sends a Ping RPC to a bootstrap node and waits for response.
func (ln *Server) PingBootstrap(bootstrapIP string, bootstrapPort int) {
	ping := &RPCMessage{
		Type:     RPCPing,
		FromID:   ln.Self.HexID(),
		FromIP:   ln.Self.ipAddr,
		FromPort: ln.Self.port,
	}

	resp, err := ln.Transport.SendRPC(bootstrapIP, bootstrapPort, ping, 5*time.Second)
	if err != nil {
		fmt.Printf("Server %s error pinging bootstrap: %v\n",
			ln.Self.HexID(), err)
		return
	}

	fmt.Printf("Server %s got Ping response: %+v\n",
		ln.Self.HexID(), resp)
}