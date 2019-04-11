package protocol

import (
	"encoding/json"
	"log"
)

type proxyAuthenticateMessageJSON struct {
	IdentityKey string `json:"identity_public_key"`
	Signature   string `json:"signature"`
}

type ProxyAuthenticateMessage struct {
	identityKey string
	signature   string
}

func NewProxyAuthenticateMessage(identityKey, signature string) *ProxyAuthenticateMessage {
	return &ProxyAuthenticateMessage{
		identityKey: identityKey,
		signature:   signature,
	}
}

func (message *ProxyAuthenticateMessage) IdentityKey() string {
	return message.identityKey
}

func (message *ProxyAuthenticateMessage) Signature() string {
	return message.signature
}

func (message ProxyAuthenticateMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&proxyAuthenticateMessageJSON{
		IdentityKey: message.IdentityKey(),
		Signature:   message.Signature(),
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "proxy_authenticate",
		Data: json.RawMessage(body),
	})
}

func (ProxyAuthenticateMessage) Type() MessageType {
	return ProxyAuthenticateMessageType
}

func decodeProxyAuthenticateMessage(data json.RawMessage) (*ProxyAuthenticateMessage, error) {
	var message proxyAuthenticateMessageJSON

	log.Printf("Decoding Proxy Authenticate message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewProxyAuthenticateMessage(message.IdentityKey, message.Signature), nil
}
