package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	isBootstrap := flag.Bool("b", false, "is boostrap node")
	port := flag.Int("p", 8090, "port number")
	bootstrapIP := flag.String("ba", "127.0.0.1", "bootstrap ip address")
	bootstrapPort := flag.Int("bp", 8090, "bootstrap node port number")

	flag.Parse()

	server, err := NewServer("127.0.0.1", *port)
	if err != nil {
		log.Fatalf("Error creating LocalNode: %v", err)
	}

	if !*isBootstrap {
		fmt.Printf("Starting JOINING node on port %d\n", *port)
		server.PingBootstrap(*bootstrapIP, *bootstrapPort)
	} else {
		fmt.Printf("Starting BOOTSTRAP node on port %d\n", *port)
	}

	// Block forever handling RPCs
	server.Run()

}
