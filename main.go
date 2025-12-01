package main

import (
	"fmt"
	//"container/heap"
	//"flag"
)

func main() {

	n1, _ := NewNodeFromIPAndPort("192.0.2.10", 4001)
	// n2, _ := NewNodeFromIPAndPort("2001:db8::1", 4001)
	// n3, _ := NewNodeFromIPAndPort("2001:db8::1", 4002)

	fmt.Println("Node 1 ID:", n1.HexID())
	// fmt.Println("Node 2 ID:", n2.HexID())
	// fmt.Println("Node 3 ID:", n3.HexID())

	// fmt.Println("Distance from N1 to N2: ", NodeIDToHex(n1.GetXorDistance(&n2)))
	// fmt.Println("Distance from N2 to N1: ", NodeIDToHex(n2.GetXorDistance(&n1)))
	// fmt.Println("Distance from N1 to N3: ", NodeIDToHex(n1.GetXorDistance(&n3)))
	// fmt.Println("Distance from N2 to N3: ", NodeIDToHex(n2.GetXorDistance(&n3)))

	// // Create a distance queue, put the nodes in it, and
	// // establish the distance queue (heap) invariants.
	// nmh := make(NodeMinHeap, 2)

	// nmh[0] = &NodeMinHeapItem{
	// 	node:    n1,
	// 	distance: n1.GetXorDistance(&n1),
	// 	index:    0,
	// }

	// nmh[1] = &NodeMinHeapItem{
	// 	node:    n3,
	// 	distance: n1.GetXorDistance(&n3),
	// 	index:    1,
	// }
	
	// heap.Init(&nmh)

	// // Insert a new NodeMinHeapItem and then modify its distance.
	// nmhi := &NodeMinHeapItem{
	// 	node:    n2,
	// 	distance: n1.GetXorDistance(&n2),
	// }
	// heap.Push(&nmh, nmhi)

	// // Take the NodeMinHeapItems out; they arrive in decreasing distance order.
	// for nmh.Len() > 0 {
	// 	nmhi := heap.Pop(&nmh).(*NodeMinHeapItem)
	// 	fmt.Printf("Node %s disance to Node 1: %s\n", NodeIDToHex(nmhi.node.node_id), NodeIDToHex(nmhi.distance))
	// }

	// isBootstrapPtr := flag.Bool("b", false, "is boostrap node")
	// portPtr := flag.Int("p", 8090, "port number")
	// bootstrapIPPtr := flag.String("ba", "127.0.0.1", "bootstrap ip address")
	// bootstrapPortPtr := flag.Int("bp", 8090, "bootstrap node port number")

	// flag.Parse()

	// if !*isBootstrapPtr {
	// 	sendUDPMessage(*bootstrapIPPtr, *bootstrapPortPtr)
	// } else {
	// 	fmt.Println("We are a bootstrap node")
	// }

	// createUDPListener(*portPtr)
}



