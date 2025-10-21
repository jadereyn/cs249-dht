package main

import "fmt"
import "github.com/go-faster/xor"

const NODE_ID_BUFFER_SIZE int = 20 // 20 bytes in 160-bit node ID

func main() {
    x := make([]byte, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)
    y := make([]byte, NODE_ID_BUFFER_SIZE, NODE_ID_BUFFER_SIZE)

    // Initializing the elements
    for i := 0; i < 10; i++ {
        x[i] = byte(i + 1)
        y[i] = byte(i + 1)
    }

    res := xor_distance(x, y)
    fmt.Println(res)
}

// assuming network byte order (big endian)
func xor_distance(x, y []byte) []byte {
    fmt.Println(x)
    fmt.Println(y)

    res := make([]byte, 20, 20)
    xor.Bytes(res, x, y)
    return res
}