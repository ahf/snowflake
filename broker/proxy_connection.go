package main

import (
	"errors"
	"log"

	"git.torproject.org/pluggable-transports/snowflake.git/common/protocol"

	"github.com/gorilla/websocket"
)

type ProxyConnection struct {
	index          uint
	connection     *websocket.Conn
	platform       string
	version        string
	messages       chan protocol.Message
	received_hello bool
	nonce          *string
	ready          bool
}

func NewProxyConnection(connection *websocket.Conn) *ProxyConnection {
	proxy_connection := &ProxyConnection{
		index:          0,
		connection:     connection,
		messages:       make(chan protocol.Message),
		received_hello: false,
		nonce:          nil,
		ready:          false,
	}

	go proxy_connection.handle_messages()

	return proxy_connection
}

func (connection *ProxyConnection) Close() error {
	return connection.connection.Close()
}

func (connection *ProxyConnection) handle_messages() {
	defer func() {
		close(connection.messages)
	}()

	for {
		_, message, err := connection.connection.ReadMessage()

		if err != nil {
			log.Printf("Unable to read WebSocket message: %s", err)
			return
		}

		log.Printf("Received message from proxy connection: %s", message)

		m, err := protocol.Decode(message)

		if err != nil {
			log.Printf("Unable to decode message: %s", err)
			return
		}

		connection.messages <- m
	}
}

func (connection *ProxyConnection) ReadMessage() (protocol.Message, error) {
	message, ok := <-connection.messages

	if !ok {
		return nil, errors.New("Proxy Message handler go-routine closed")
	}

	return message, nil
}

func (connection *ProxyConnection) WriteMessage(message protocol.Message) error {
	data, err := message.Encode()

	if err != nil {
		return nil
	}

	log.Printf("Sending message to proxy: %s", data)
	err = connection.connection.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		return err
	}

	return nil
}

func (connection *ProxyConnection) Handle(ctx *BrokerContext) error {
	defer func() {
		connection.Close()
	}()

	for {
		message, err := connection.ReadMessage()

		if err != nil {
			return err
		}

		// Better safe than sorry?
		if !connection.received_hello && message.Type() != protocol.ProxyHelloMessageType {
			return errors.New("Protocol violation: First message must be Proxy Hello message")
		}

		switch message.Type() {
		case protocol.ProxyHelloMessageType:
			if connection.received_hello {
				log.Printf("Received repeated Proxy Hello message. Closing connection")
				return errors.New("Protocol violation: Repeated Proxy Hello message")
			}

			connection.received_hello = true
			connection.platform = message.(*protocol.ProxyHelloMessage).Platform()
			connection.version = message.(*protocol.ProxyHelloMessage).Version()

			log.Printf("Platform: %s (%s)", connection.platform, connection.version)

			// Used to authenticate proxies.
			broker_hello := protocol.NewDefaultBrokerHelloMessage()

			nonce := broker_hello.Nonce()
			connection.nonce = &nonce

			// Write Broker Hello message to the proxy
			err := connection.WriteMessage(broker_hello)

			if err != nil {
				return err
			}

		case protocol.ProxyReadyMessageType:
			if !connection.received_hello {
				return errors.New("Protocol violation: Proxy marked itself ready before hello message.")
			}
		}
	}

	return nil
}

func (connection *ProxyConnection) SetConnectionSetIndex(index uint) {
	connection.index = index
}

func (connection *ProxyConnection) ConnectionSetIndex() uint {
	return connection.index
}
