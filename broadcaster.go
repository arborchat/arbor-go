package main

import (
	"io"
	"log"
	"sync"
)

type Broadcaster struct {
	sync.RWMutex
	clients map[io.ReadWriteCloser]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[io.ReadWriteCloser]struct{}),
	}
}

func (b *Broadcaster) Send(data []byte) {
	b.RLock()
	for client := range b.clients {
		_, err := client.Write(data)
		if err != nil {
			log.Println("err writing to client, removing", err)
			go b.remove(client)
		}
	}
	b.RUnlock()
}

func (b *Broadcaster) remove(client io.ReadWriteCloser) {
	b.Lock()
	delete(b.clients, client)
	b.Unlock()
}

func (b *Broadcaster) Add(client io.ReadWriteCloser) {
	b.Lock()
	b.clients[client] = struct{}{}
	b.Unlock()
}
