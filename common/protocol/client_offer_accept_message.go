package protocol

import (
	"encoding/json"
	"log"
)

type clientOfferAcceptMessageJSON struct {
	Descriptor string `json:"descriptor"`
}

type ClientOfferAcceptMessage struct {
	descriptor string
}

func NewClientOfferAcceptMessage(descriptor string) *ClientOfferAcceptMessage {
	return &ClientOfferAcceptMessage{
		descriptor: descriptor,
	}
}

func (message ClientOfferAcceptMessage) Encode() ([]byte, error) {
	body, err := json.Marshal(&clientOfferAcceptMessageJSON{
		Descriptor: message.Descriptor(),
	})

	if err != nil {
		return nil, err
	}

	return json.Marshal(rawMessage{
		Type: "client_offer_accept",
		Data: json.RawMessage(body),
	})
}

func (ClientOfferAcceptMessage) Type() MessageType {
	return ClientOfferAcceptMessageType
}

func (message *ClientOfferAcceptMessage) Descriptor() string {
	return message.descriptor
}

func decodeClientOfferAcceptMessage(data json.RawMessage) (*ClientOfferAcceptMessage, error) {
	var message clientOfferAcceptMessageJSON

	log.Printf("Decoding Client Offer Accept message body: %s", string(data))

	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return NewClientOfferAcceptMessage(message.Descriptor), nil
}
