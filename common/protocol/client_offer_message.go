package protocol

import (
	"encoding/json"
	"log"
)

type clientOfferMessageJSON struct {
	Descriptor string `json:"descriptor"`
}

type ClientOfferMessage struct {
	descriptor string
}

func NewClientOfferMessage(descriptor string) *ClientOfferMessage {
	return &ClientOfferMessage{
		descriptor: descriptor,
	}
}

func (message ClientOfferMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&clientOfferMessageJSON{
		Descriptor: message.Descriptor(),
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "client_offer",
		Data: json.RawMessage(body),
	})
}

func (ClientOfferMessage) Type() MessageType {
	return ClientOfferMessageType
}

func (message *ClientOfferMessage) Descriptor() string {
	return message.descriptor
}

func decodeClientOfferMessage(data json.RawMessage) (*ClientOfferMessage, error) {
	var message clientOfferMessageJSON

	log.Printf("Decoding Client Offer message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewClientOfferMessage(message.Descriptor), nil
}
