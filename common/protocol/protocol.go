package protocol

import (
	"encoding/json"
	"errors"
	"log"
)

type Message interface {
	Type() MessageType
	Encode() ([]byte, error)
}

type rawMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func Decode(data []byte) (Message, error) {
	var message rawMessage

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	log.Printf("Decoding message type %s: %s", message.Type, string(data))

	switch message.Type {
	case "proxy_hello":
		return decodeProxyHelloMessage(message.Data)
	case "broker_hello":
		return decodeBrokerHelloMessage(message.Data)
	case "client_offer":
		return decodeClientOfferMessage(message.Data)
	case "client_offer_accept":
		return decodeClientOfferAcceptMessage(message.Data)
	case "proxy_ready":
		return decodeProxyReadyMessage(message.Data)
	case "proxy_authenticate":
		return decodeProxyAuthenticateMessage(message.Data)
	}

	return nil, errors.New("Unknown message")
}
