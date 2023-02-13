package peerconnect

const (
	Choke         = uint8(0)
	Unchoke       = uint8(1)
	Interested    = uint8(2)
	NotInterested = uint8(3)
	Have          = uint8(4)
	BitField      = uint8(5)
	Request       = uint8(6)
	Piece         = uint8(7)
	Cancel        = uint8(8)
)

type Message struct {
	ID      uint8
	Payload []byte
}

// Have, Piece anad BitField Msg
type SpecialMsg struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

func ChokeMsg() []byte {
	msg := Message{ID: Choke}

	return msg.BufferMsg()
}

func UnchokeMsg() []byte {
	msg := Message{ID: Unchoke}

	return msg.BufferMsg()
}

func RequestMsg(sp *SpecialMsg) []byte {
	msg := RequestMsgPayload(sp)

	return msg.BufferMsg()
}

func InterestedMsg() []byte {
	msg := Message{ID: Interested}

	return msg.BufferMsg()
}

func NotInterestedMsg() []byte {
	msg := Message{ID: NotInterested}

	return msg.BufferMsg()
}

func HaveMsg(index uint32) []byte {
	msg := HaveMsgPayload(index)

	return msg.BufferMsg()
}
