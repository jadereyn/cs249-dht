package main

import (
	"fmt"
	"net"
	"time"
)

func sendUDPMessage(addr string, port int) {
	addrs := fmt.Sprintf("%s:%d", addr, port)
	addr2, err := net.ResolveUDPAddr("udp", addrs)
	if err != nil {
        fmt.Println("Error resolving UDP address:", err)
        return
    }

	s := fmt.Sprintf("I am sending %s:%d a message over UDP", addr, port)

	conn, err := net.DialUDP("udp", nil, addr2)
	if err != nil {
        fmt.Println("Error dialing node:", err)
        return
    }
    defer conn.Close()

    _, err = conn.Write([]byte(s))
    if err != nil {
        fmt.Printf("Error sending message %v", err)
    }

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    buffer := make([]byte, 1024)
    n, _, err := conn.ReadFromUDP(buffer)
    if err != nil {
        fmt.Printf("Error reading UDP packet: %v", err)
        return
    }
    fmt.Printf("Server says: %s\n", string(buffer[:n]))
}

func sendUDPResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	s := fmt.Sprintf("I am sending %s:%d a response over UDP", addr.IP.String(), addr.Port)
    _, err := conn.WriteToUDP([]byte(s), addr)
    if err != nil {
        fmt.Printf("Error sending message %v", err)
    }
}


func createUDPListener(port int) {
    p := make([]byte, 2048)

    addr := net.UDPAddr{
        Port: port,
        IP: net.ParseIP("127.0.0.1"),
    }

	fmt.Printf("Starting UDP listener on port %d...\n", port)
    srv, err := net.ListenUDP("udp", &addr)

    if err != nil {
        fmt.Printf("Some error %v\n", err)
        return
    }

    for {
        _,remoteaddr,err := srv.ReadFromUDP(p)
        fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
        if err !=  nil {
            fmt.Printf("Some error  %v", err)
            continue
        }

		fmt.Println("Sending response...")
        sendUDPResponse(srv, remoteaddr)
    }
}