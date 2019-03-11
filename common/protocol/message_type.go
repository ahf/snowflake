package protocol

type MessageType int

const (
	ProxyHelloMessageType = iota
	BrokerHelloMessageType
	ClientOfferMessageType
	ClientOfferAcceptMessageType
	ProxyReadyMessageType
	UnknownMessageType
)

func (message_type MessageType) String() string {
	switch message_type {
	case ProxyHelloMessageType:
		return "Proxy Hello"
	case BrokerHelloMessageType:
		return "Broker Hello"
	case ClientOfferMessageType:
		return "Client Offer"
	case ClientOfferAcceptMessageType:
		return "Client Offer Accept"
	case ProxyReadyMessageType:
		return "Proxy Ready"
	case UnknownMessageType:
		return "Unknown"
	}

	panic("Unknown message type")
	return ""
}
