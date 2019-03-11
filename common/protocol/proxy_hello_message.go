package protocol

import (
	"encoding/json"
	"log"
)

type proxyHelloMessageJSON struct {
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

type ProxyHelloMessage struct {
	platform string
	version  string
}

func NewProxyHelloMessage(platform, version string) *ProxyHelloMessage {
	return &ProxyHelloMessage{
		platform: platform,
		version:  version,
	}
}

func (message ProxyHelloMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&proxyHelloMessageJSON{
		Platform: message.Platform(),
		Version:  message.Version(),
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "proxy_hello",
		Data: json.RawMessage(body),
	})
}

func (ProxyHelloMessage) Type() MessageType {
	return ProxyHelloMessageType
}

func (message *ProxyHelloMessage) Platform() string {
	return message.platform
}

func (message *ProxyHelloMessage) Version() string {
	return message.version
}

func decodeProxyHelloMessage(data json.RawMessage) (*ProxyHelloMessage, error) {
	var message proxyHelloMessageJSON

	log.Printf("Decoding Proxy Hello message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewProxyHelloMessage(message.Platform, message.Version), nil
}
