package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"log"
	"net/url"

	"git.torproject.org/pluggable-transports/snowflake.git/common/protocol"

	"github.com/gorilla/websocket"
)

type BrokerConnection struct {
	connection     *websocket.Conn
	received_hello bool
	messages       chan protocol.Message
	secretKey      *ecdsa.PrivateKey
}

func DialBroker(url *url.URL, platform, version string, secretKey *ecdsa.PrivateKey) (*BrokerConnection, error) {
	connection, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		return nil, err
	}

	log.Printf("Connected to broker at %s", url)

	broker_connection := &BrokerConnection{
		connection:     connection,
		received_hello: false,
		messages:       make(chan protocol.Message),
		secretKey:      secretKey,
	}

	// Send the Proxy Hello message.
	broker_connection.WriteMessage(protocol.NewProxyHelloMessage(platform, version))

	// Spawn reader Goroutine.
	go broker_connection.handle_messages()

	return broker_connection, nil
}

func (conn *BrokerConnection) WriteMessage(message protocol.Message) error {
	data, err := message.Encode()

	if err != nil {
		return err
	}

	log.Printf("Sending %s", data)

	err = conn.connection.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		return err
	}

	return nil
}

func (conn *BrokerConnection) Handle() {
	for {
		select {
		case message, ok := <-conn.messages:
			if !ok {
				// FIXME(ahf): Handle.
				return
			}

			err := conn.HandleMessage(message)

			if err != nil {
				return
			}
		}
	}
}

func (conn *BrokerConnection) HandleMessage(message protocol.Message) error {
	switch message.Type() {
	case protocol.BrokerHelloMessageType:
		if conn.received_hello {
			return errors.New("Protocol violation: Received repeated Broker Hello message")
		}

		conn.received_hello = true

		// Sign the nonce with our secretKey and authenticate.
		nonce := message.(*protocol.BrokerHelloMessage).Nonce()

		nonceHash := sha256.New()
		nonceHash.Write([]byte(nonce))
		nonceHashSum := nonceHash.Sum(nil)

		signature, err := conn.secretKey.Sign(rand.Reader, nonceHashSum, nil)

		if err != nil {
			log.Printf("Unable to generate signature and authenticate: %s", err)
			return err
		}

		publicKey, err := x509.MarshalPKIXPublicKey(conn.secretKey.Public())

		if err != nil {
			log.Printf("Unable to encode public key: %s", err)
			return err
		}

		publicKeyEncoded := base64.StdEncoding.EncodeToString(publicKey)
		signatureEncoded := base64.StdEncoding.EncodeToString(signature)

		conn.WriteMessage(protocol.NewProxyAuthenticateMessage(publicKeyEncoded, signatureEncoded))

	}

	return nil
}

func (conn *BrokerConnection) handle_messages() {
	defer func() {
		close(conn.messages)
	}()

	for {
		_, data, err := conn.connection.ReadMessage()

		if err != nil {
			log.Printf("Unable to read message from broker: %s", err)
			return
		}

		log.Printf("Received message from broker: %s", data)

		message, err := protocol.Decode(data)

		if err != nil {
			log.Printf("Invalid message from broker: %s", err)
		}

		conn.messages <- message
	}
}
