package main

import (
	"testing"
	"net"
	"fmt"
)

func TestAppServer(t *testing.T) {
	conn, _ := net.ListenPacket("udp", ":8999")
	buff := make([]byte, 64 * 1024)
	for {
		n, _, _ := conn.ReadFrom(buff)
		fmt.Println(buff[:n])
	}
}
