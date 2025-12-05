package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
)

func main() {

	isBootstrap := flag.Bool("b", false, "is boostrap node")
	port := flag.Int("p", 8090, "port number")
	bootstrapIP := flag.String("ba", "127.0.0.1", "bootstrap ip address")
	bootstrapPort := flag.Int("bp", 8090, "bootstrap node port number")

	lookupTargetHex := flag.String("lookup", "", "hex node ID to lookup")

	flag.Parse()

	server, err := NewServer("127.0.0.1", *port)
	if err != nil {
		log.Fatalf("Error creating LocalNode: %v", err)
	}

	if !*isBootstrap {
		fmt.Printf("Starting JOINING node on port %d\n", *port)
		server.PingBootstrap(*bootstrapIP, *bootstrapPort)
		fmt.Printf("After PingBootstrap, router has %d buckets, bucket[0] size=%d\n",
			len(server.Router.buckets),
			server.Router.buckets[0].Len(),
		)

		// targetID := ln.Self.nodeID // e.g. lookup our own ID as a test
		// nodes, err := ln.LookupNodes(targetID)

		// If user requested a lookup, do one
		if *lookupTargetHex != "" {
			targetID := new(big.Int)
			if _, ok := targetID.SetString(*lookupTargetHex, 16); !ok {
				log.Fatalf("invalid lookup target hex: %s", *lookupTargetHex)
			}

			fmt.Printf("Running LookupNodes for targetID=%s...\n", *lookupTargetHex)
			nodes, err := server.LookupNodes(targetID)
			if err != nil {
				fmt.Printf("LookupNodes error: %v\n", err)
			} else {
				fmt.Println("LookupNodes returned:")
				for _, n := range nodes {
					fmt.Printf("- %s:%d id=%s\n", n.ipAddr, n.port, n.HexID())
				}
			}
		}
	} else {
		fmt.Printf("Starting BOOTSTRAP node on port %d\n", *port)
	}

	// Block forever handling RPCs
	server.Run()

}
