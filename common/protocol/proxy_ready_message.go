package protocol

import (
	"encoding/json"
	"log"
)

type proxyReadyMessageJSON struct {
}

type ProxyReadyMessage struct {
	// FIXME(ahf): Should we include the number of clients (capacity) we assume
	// we can handle here?
}

func NewProxyReadyMessage() *ProxyReadyMessage {
	return &ProxyReadyMessage{}
}

func (message ProxyReadyMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&proxyReadyMessageJSON{})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "proxy_ready",
		Data: json.RawMessage(body),
	})
}

func (ProxyReadyMessage) Type() MessageType {
	return ProxyReadyMessageType
}

func decodeProxyReadyMessage(data json.RawMessage) (*ProxyReadyMessage, error) {
	var message proxyReadyMessageJSON

	log.Printf("Decoding Proxy Ready message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewProxyReadyMessage(), nil
}
