package main

//"fmt"
//"encoding/json"

type RPCDescriptor int

const (
	RPCPing RPCDescriptor = iota
	RPCPong
	RPCFindNode
	RPCFindNodeResp
	RPCStore
	RPCFindValue
	RPCFindValueResp
)

var stateName = map[RPCDescriptor]string{
	RPCPing:      "Ping",
	RPCPong:      "Pong",
	RPCFindNode:  "Find Node",
	RPCFindNodeResp:  "Find Node Response",
	RPCStore:     "Store",
	RPCFindValue: "Find Value",
	RPCFindValueResp: "Find Value Response",
}

// Node info that we send over the wire (simplified)
type RPCNodeInfo struct {
	ID   string `json:"id"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// RPCMessage is what we send over the wire as JSON.
type RPCMessage struct {
	Type     RPCDescriptor `json:"type"`
	FromID   string        `json:"from_id"` // hex node id, or dummy for now
	FromIP   string        `json:"from_ip"`
	FromPort int           `json:"from_port"`

	// For FIND_NODE / FIND_VALUE
	TargetID string        `json:"target_id,omitempty"`
	Nodes    []RPCNodeInfo `json:"nodes,omitempty"`

	// For STORE / FIND_VALUE
	Key   string `json:"key,omitempty"`
	Value []byte `json:"value,omitempty"`
}
