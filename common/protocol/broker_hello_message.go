package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

const ProtocolVersion = "1"

type brokerHelloMessageJSON struct {
	Protocol  string `json:"protocol"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
}

type BrokerHelloMessage struct {
	protocol  string
	timestamp string
	nonce     string
}

func NewBrokerHelloMessage(protocol, timestamp, nonce string) *BrokerHelloMessage {
	return &BrokerHelloMessage{
		protocol:  protocol,
		timestamp: timestamp,
		nonce:     nonce,
	}
}

func NewDefaultBrokerHelloMessage() *BrokerHelloMessage {
	timestamp := time.Now().Format(time.RFC3339)
	nonce := newNonce()
	return NewBrokerHelloMessage(ProtocolVersion, timestamp, nonce)
}

func (message BrokerHelloMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&brokerHelloMessageJSON{
		Protocol:  message.Protocol(),
		Timestamp: message.Timestamp(),
		Nonce:     message.Nonce(),
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "broker_hello",
		Data: json.RawMessage(body),
	})
}

func (BrokerHelloMessage) Type() MessageType {
	return BrokerHelloMessageType
}

func (message *BrokerHelloMessage) Protocol() string {
	return message.protocol
}

func (message *BrokerHelloMessage) Timestamp() string {
	return message.timestamp
}

func (message *BrokerHelloMessage) Nonce() string {
	return message.nonce
}

func decodeBrokerHelloMessage(data json.RawMessage) (*BrokerHelloMessage, error) {
	var message brokerHelloMessageJSON

	log.Printf("Decoding Broker Hello message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewBrokerHelloMessage(message.Protocol, message.Timestamp, message.Nonce), nil
}

func newNonce() string {
	const count = 32
	buffer := make([]byte, count)

	_, err := rand.Read(buffer)

	if err != nil {
		log.Printf("Unable to generate random nonce")

		// FIXME(ahf): What should we do here?
	}

	h := sha256.New()
	h.Write(buffer)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
