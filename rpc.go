package main

import (
	//"fmt"
	//"encoding/json"
)

type RPCDescriptor int

const (
    RPCPing RPCDescriptor = iota
    RPCFindNode
    RPCStore
    RPCFindValue
)

var stateName = map[RPCDescriptor]string {
    RPCPing:      "Ping",
    RPCFindNode: "Find Node",
    RPCStore:     "Store",
    RPCFindValue:  "Find Value",
}

type RPCMessage struct {
	descriptor RPCDescriptor
}