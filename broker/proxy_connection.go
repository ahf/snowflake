package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"math/big"

	"git.torproject.org/pluggable-transports/snowflake.git/common/protocol"

	"github.com/gorilla/websocket"
)

// Should be kept in sync with Go's ecdsa.go? Wat.
type ecdsaSignature struct {
	R, S *big.Int
}

type ProxyConnection struct {
	index          uint
	connection     *websocket.Conn
	platform       string
	protocol       string
	version        string
	messages       chan protocol.Message
	received_hello bool
	nonce          *string
	ready          bool
	publicKey      *ecdsa.PublicKey
}

func NewProxyConnection(connection *websocket.Conn) *ProxyConnection {
	proxy_connection := &ProxyConnection{
		index:          0,
		connection:     connection,
		messages:       make(chan protocol.Message),
		received_hello: false,
		nonce:          nil,
		ready:          false,
		publicKey:      nil,
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
			connection.protocol = message.(*protocol.ProxyHelloMessage).Protocol()

			log.Printf("Platform: %s (%s) using protocol v%s", connection.platform, connection.version, connection.protocol)

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

		case protocol.ProxyAuthenticateMessageType:
			if !connection.received_hello {
				return errors.New("Protocol violation: Proxy wanted to authenticate before hello message.")
			}

			if connection.nonce == nil {
				// FIXME(ahf): Should never happen, but the nonce code could be prettier.
				return errors.New("Protocol violation: Strange, we do not seem to have a nonce?")
			}

			// Extract identity public key.
			identityKeyEncoded := message.(*protocol.ProxyAuthenticateMessage).IdentityKey()
			identityKeyDER, err := base64.StdEncoding.DecodeString(identityKeyEncoded)

			if err != nil {
				return err
			}

			identityKey, err := x509.ParsePKIXPublicKey(identityKeyDER)

			if err != nil {
				return err
			}

			// Extract signature.
			signatureEncoded := message.(*protocol.ProxyAuthenticateMessage).Signature()
			signatureDER, err := base64.StdEncoding.DecodeString(signatureEncoded)

			if err != nil {
				return nil
			}

			signature := &ecdsaSignature{}

			_, err = asn1.Unmarshal(signatureDER, signature)

			if err != nil {
				return err
			}

			// Hash our nonce.
			nonceHash := sha256.New()
			nonceHash.Write([]byte(*connection.nonce))
			nonceHashSum := nonceHash.Sum(nil)

			// Check out signature.
			valid := ecdsa.Verify(identityKey.(*ecdsa.PublicKey), nonceHashSum, signature.R, signature.S)

			if valid {
				h := sha256.New()
				h.Write(identityKeyDER)
				fingerprint := hex.EncodeToString(h.Sum(nil))

				log.Printf("Proxy %s have successfully authenticated", fingerprint)

				connection.publicKey = identityKey.(*ecdsa.PublicKey)
			} else {
				log.Printf("Proxy failed to authenticated")
				return errors.New("Proxy authentication failed")
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