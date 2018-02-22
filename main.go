package main

import (
	"encoding/json"
	"io"
	"log"
	"net"

	. "github.com/whereswaldon/arbor/messages"
)

func main() {
	messages := NewStore()
	broadcaster := NewBroadcaster()
	//serve
	listener, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on localhost:7777")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		broadcaster.Add(conn)
		go handleClient(conn, messages, broadcaster)
		m, err := NewMessage("A new client has joined")
		if err != nil {
			log.Println(err)
		}
		a := &ArborMessage{
			Type:    NEW_MESSAGE,
			Message: m,
		}
		go handleNewMessage(a, messages, broadcaster)
	}
}

func handleClient(conn io.ReadWriteCloser, store *Store, broadcaster *Broadcaster) {
	data := make([]byte, 1024)
	for {
		_, err := conn.Read(data)
		if err != nil {
			log.Println(err)
			return
		}
		a := &ArborMessage{}
		err = json.Unmarshal(data, a)
		if err != nil {
			log.Println(err)
			continue
		}
		switch a.Type {
		case QUERY:
			go handleQuery(a, conn, store)
		case NEW_MESSAGE:
			go handleNewMessage(a, store, broadcaster)
		default:
			log.Println("Unrecognized message type", a.Type)
			continue
		}
	}
}

func handleQuery(msg *ArborMessage, conn io.ReadWriteCloser, store *Store) {
	result := store.Get(msg.Message.UUID)
	msg.Message = result
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling response", err)
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Println("Error sending response", err)
	}
}

func handleNewMessage(msg *ArborMessage, store *Store, broadcaster *Broadcaster) {
	store.Add(msg.Message)
	j, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
	}
	broadcaster.Send(j)
}
