package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// UDPTransport owns a single UDP socket that a node uses
// for both sending and receiving messages.
type UDPTransport struct {
	conn *net.UDPConn // underlying socket
	addr *net.UDPAddr // local address (IP + port)
}

// NewUDPTransport creates a UDP socket bound to listenIP:port.
// This will be the single socket used for both send + receive.
func NewUDPTransport(listenIP string, port int) (*UDPTransport, error) {
	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(listenIP),
		Port: port,
	}

	fmt.Printf("Binding UDP socket on %s:%d\n", listenIP, port)
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("listen udp: %w", err)
	}

	return &UDPTransport{
		conn: conn,
		addr: localAddr,
	}, nil
}

// Listen starts a loop that receives packets and passes them to a handler.
func (t *UDPTransport) Listen(handler func(data []byte, from *net.UDPAddr)) {
	buf := make([]byte, 2048)

	fmt.Printf("Starting UDP listener on %s:%d...\n", t.addr.IP.String(), t.addr.Port)

	for {
		n, remoteAddr, err := t.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Error reading UDP packet: %v\n", err)
			continue
		}

		// Make a copy of the data slice for safety
		data := make([]byte, n)
		copy(data, buf[:n])

		// Call user-provided handler
		handler(data, remoteAddr)
	}
}

// SendRPC sends an RPCMessage as JSON to addr:port and waits for a single response.
func (t *UDPTransport) SendRPC(addr string, port int, msg *RPCMessage, timeout time.Duration) (*RPCMessage, error) {
	remoteStr := fmt.Sprintf("%s:%d", addr, port)
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteStr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp addr: %w", err)
	}

	// Marshal the RPCMessage to JSON
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal rpc: %w", err)
	}

	// Ensure we clear any read deadline when we return
	defer t.conn.SetReadDeadline(time.Time{})

	// Send bytes via our single shared socket
	if _, err := t.conn.WriteToUDP(payload, remoteAddr); err != nil {
		return nil, fmt.Errorf("write to udp: %w", err)
	}

	// Set a short-lived read deadline just for this RPC
	if err := t.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("set read deadline: %w", err)
	}

	// Read one response packet
	buf := make([]byte, 2048)
	n, _, err := t.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, fmt.Errorf("read from udp: %w", err)
	}

	var resp RPCMessage
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return nil, fmt.Errorf("unmarshal rpc response: %w", err)
	}

	return &resp, nil
}

// func (t *UDPTransport) sendUDPMessage(addr string, port int, msg string, timeout time.Duration) (string, error) {
// 	addrs := fmt.Sprintf("%s:%d", addr, port)
// 	addr2, err := net.ResolveUDPAddr("udp", addrs)
// 	if err != nil {
// 		fmt.Println("Error resolving UDP address:", err)
// 		return "", fmt.Errorf("resolve udp addr: %w", err)
// 	}

// 	// s := fmt.Sprintf("I am sending %s:%d a message over UDP", addr, port)

// 	// conn, err := net.DialUDP("udp", nil, addr2)
// 	// if err != nil {
// 	// 	fmt.Println("Error dialing node:", err)
// 	// 	return
// 	// }
// 	// defer conn.Close()

// 	defer t.conn.SetReadDeadline(time.Time{})

// 	// Send the message out our single shared socket.
// 	_, err = t.conn.WriteToUDP([]byte(msg), addr2)
// 	if err != nil {
// 		fmt.Printf("Error sending message %v", err)
// 		return "", fmt.Errorf("write to udp: %w", err)
// 	}

// 	// Set a short-lived read deadline JUST for this RPC.
// 	if err := t.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
// 		return "", fmt.Errorf("set read deadline: %w", err)
// 	}

// 	buffer := make([]byte, 1024)
// 	n, _, err := t.conn.ReadFromUDP(buffer)
// 	if err != nil {
// 		fmt.Printf("Error reading UDP packet: %v", err)
// 		return "", fmt.Errorf("read from udp: %w", err)
// 	}

// 	fmt.Printf("Server says: %s\n", string(buffer[:n]))
// 	return string(buffer[:n]), nil
// }

// func sendUDPResponse(conn *net.UDPConn, addr *net.UDPAddr) {
// 	s := fmt.Sprintf("I am sending %s:%d a response over UDP", addr.IP.String(), addr.Port)
// 	_, err := conn.WriteToUDP([]byte(s), addr)
// 	if err != nil {
// 		fmt.Printf("Error sending message %v", err)
// 	}
// }

// func createUDPListener(port int) {
// 	fmt.Printf("Creating UDP listener on port %d\n", port)
// 	p := make([]byte, 2048)

// 	addr := net.UDPAddr{
// 		Port: port,
// 		IP:   net.ParseIP("127.0.0.1"),
// 	}

// 	fmt.Printf("Starting UDP listener on port %d...\n", port)
// 	srv, err := net.ListenUDP("udp", &addr)

// 	if err != nil {
// 		fmt.Printf("Some error %v\n", err)
// 		return
// 	}

// 	for {
// 		_, remoteaddr, err := srv.ReadFromUDP(p)
// 		fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
// 		if err != nil {
// 			fmt.Printf("Some error  %v", err)
// 			continue
// 		}

// 		fmt.Println("Sending response...")
// 		sendUDPResponse(srv, remoteaddr)
// 	}
// }
