package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"time"

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
	go func() {
		m, err := NewMessage("Root message")
		if err != nil {
			log.Println(err)
		}
		log.Println("Root message UUID is " + m.UUID)
		messages.Add(m)
		for t := range time.NewTicker(time.Second).C {
			m, err = m.Reply("It is now " + t.String())
			if err != nil {
				log.Println(err)
			}
			a := &ArborMessage{
				Type:    NEW_MESSAGE,
				Message: m,
			}
			go handleNewMessage(a, messages, broadcaster)
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		broadcaster.Add(conn)
		go handleClient(conn, messages, broadcaster)
	}
}

func handleClient(conn io.ReadWriteCloser, store *Store, broadcaster *Broadcaster) {
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			log.Println(err)
			return
		}
		a := &ArborMessage{}
		err = json.Unmarshal(data[:n], a)
		if err != nil {
			log.Println(err)
			continue
		}
		switch a.Type {
		case QUERY:
			log.Println("Handling query for " + a.Message.UUID)
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
	if result == nil {
		log.Println("Unable to find queried id: " + msg.Message.UUID)
		return
	}
	msg.Message = result
	msg.Type = NEW_MESSAGE
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling response", err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Println("Error sending response", err)
		return
	}
	log.Println("Query response: ", string(data))
}

func handleNewMessage(msg *ArborMessage, store *Store, broadcaster *Broadcaster) {
	store.Add(msg.Message)
	j, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
	}
	broadcaster.Send(j)
}
