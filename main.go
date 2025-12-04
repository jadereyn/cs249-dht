package main

import (
	"flag"
	"fmt"
)

func main() {

	isBootstrapPtr := flag.Bool("b", false, "is boostrap node")
	portPtr := flag.Int("p", 8090, "port number")
	bootstrapIPPtr := flag.String("ba", "127.0.0.1", "bootstrap ip address")
	bootstrapportPtr := flag.Int("bp", 8090, "bootstrap node port number")

	flag.Parse()

	if !*isBootstrapPtr {
		sendUDPMessage(*bootstrapIPPtr, *bootstrapportPtr)
	} else {
		fmt.Println("We are a bootstrap node")
	}

	createUDPListener(*portPtr)
}
