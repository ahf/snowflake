package main

import (
	"log"
)

type ProxyConnectionSet struct {
	counter     uint
	connections map[uint]*ProxyConnection
}

func NewProxyConnectionSet() *ProxyConnectionSet {
	return &ProxyConnectionSet{
		counter:     1,
		connections: make(map[uint]*ProxyConnection),
	}
}

func (set *ProxyConnectionSet) Add(connection *ProxyConnection) {
	set.counter = set.counter + 1
	counter := set.counter
	connection.SetConnectionSetIndex(counter)
	log.Printf("Registering proxy connection %d", counter)
	set.connections[counter] = connection
}

func (set *ProxyConnectionSet) Remove(connection *ProxyConnection) {
	log.Printf("Unregistering proxy connection %d", connection.ConnectionSetIndex())
	delete(set.connections, connection.ConnectionSetIndex())
}
