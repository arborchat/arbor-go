package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
)

type ArborMessageType uint8

const (
	QUERY          = 0
	NEW_MESSAGE    = 1
	CREATE_MESSAGE = 2
)

type ArborMessage struct {
	Type ArborMessageType
	*Message
}

func main() {
	if len(os.Args) > 1 {
		//connect to arg 1
		conn, err := net.Dial("tcp", os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		io.Copy(os.Stdout, conn)
	} else {
		//messages := NewStore()
		broadcaster := NewBroadcaster()
		//serve
		listener, err := net.Listen("tcp", ":7777")
		if err != nil {
			log.Fatal(err)
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
			}
			broadcaster.Add(conn)
			m, err := NewMessage("A new client has joined")
			if err != nil {
				log.Println(err)
			}
			a := &ArborMessage{
				Type:    NEW_MESSAGE,
				Message: m,
			}
			j, err := json.Marshal(a)
			if err != nil {
				log.Println(err)
			}
			broadcaster.Send(j)
		}
	}
}
