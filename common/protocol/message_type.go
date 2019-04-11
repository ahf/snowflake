package protocol

type MessageType int

const (
	ProxyHelloMessageType MessageType = iota
	BrokerHelloMessageType
	ClientOfferMessageType
	ClientOfferAcceptMessageType
	ProxyReadyMessageType
	ProxyAuthenticateMessageType
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
	case ProxyAuthenticateMessageType:
		return "Proxy Authenticate"
	case UnknownMessageType:
		return "Unknown"
	}

	panic("Unknown message type")
	return ""
}
