package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <host:port>")
		return
	}
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		fmt.Println("Unable to connect", err)
		return
	}
	io.Copy(os.Stdout, conn)
}
