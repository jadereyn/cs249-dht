package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {

	isBootstrapPtr := flag.Bool("b", false, "is boostrap node")
	portPtr := flag.Int("p", 8090, "port number")
	bootstrapIPPtr := flag.String("ba", "127.0.0.1", "bootstrap ip address")
	bootstrapPortPtr := flag.Int("bp", 8090, "bootstrap node port number")

	flag.Parse()

	transport, err := NewUDPTransport("127.0.0.1", *portPtr)
	if err != nil {
		log.Fatal("Error creating UDP transport:", err)
	}

	// if err != nil {
	// 	fmt.Printf("Error creating UDP transport: %v\n", err)
	// 	return
	// }

	selfIP := transport.addr.IP.String()
	selfPort := transport.addr.Port

	if !*isBootstrapPtr {
		// We are a joining node: send a Ping to the bootstrap node
		ping := &RPCMessage{
			Type:     RPCPing,
			FromID:   "dummy-id", // later: NodeIDToHex(self.node_id)
			FromIP:   selfIP,
			FromPort: selfPort,
		}

		resp, err := transport.SendRPC(*bootstrapIPPtr, *bootstrapPortPtr, ping, 5*time.Second)
		if err != nil {
			fmt.Printf("Error talking to bootstrap: %v\n", err)
		} else {
			fmt.Printf("Bootstrap replied with RPC: %+v\n", resp)
		}
	} else {
		fmt.Println("We are a bootstrap node")
	}

	// Start listening loop (blocks forever)
	transport.Listen(func(data []byte, from *net.UDPAddr) {
		var msg RPCMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			fmt.Printf("Error unmarshaling RPCMessage: %v\n", err)
			return
		}
		fmt.Printf("Got RPC %v from %v\n", msg.Type, from)

		switch msg.Type {
		case RPCPing:
			// Reply with Pong
			pong := &RPCMessage{
				Type:     RPCPong,
				FromID:   "bootstrap-id",
				FromIP:   selfIP,
				FromPort: selfPort,
			}
			payload, err := json.Marshal(pong)
			if err != nil {
				fmt.Printf("Error marshaling pong: %v\n", err)
				return
			}
			if _, err := transport.conn.WriteToUDP(payload, from); err != nil {
				fmt.Printf("Error sending pong: %v\n", err)
			}

		case RPCPong:
			// For now, just log it.
			fmt.Printf("Received Pong from %s:%d\n", msg.FromIP, msg.FromPort)

		default:
			fmt.Printf("Unknown RPC type: %v\n", msg.Type)
		}
	})

	// createUDPListener(*portPtr)
}
