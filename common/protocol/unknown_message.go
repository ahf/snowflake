package protocol

type UnknownMessage struct {
}

func (UnknownMessage) Type() MessageType {
	return UnknownMessageType
}
