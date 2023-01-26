package peerconnection

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

const (
	Choke         = uint8(0)
	UnChoke       = uint8(1)
	Intereseted   = uint8(2)
	NotInterested = uint8(3)
	Have          = uint8(4)
	Bitfield      = uint8(5)
	Request       = uint8(6)
	Piece         = uint8(7)
	Cancel        = uint8(8)
)

type Message struct {
	ID      uint8
	Payload []byte
}

// Converts a message into a buffer of bytes in the form of
// length.messageID.payload
// Refer to BEP 003 specification
func (msg *Message) BufferMessage() []byte {
	// null messages are keep-alive messages
	if msg == nil {
		return make([]byte, 4)
	}

	msgLength := uint32(len(msg.Payload) + 1)
	msgBuffer := make([]byte, 4+msgLength)

	binary.BigEndian.PutUint32(msgBuffer[0:4], msgLength)
	msgBuffer[4] = byte(msg.ID)
	copy(msgBuffer[5:], msg.Payload)

	return msgBuffer
}

// Parses incoming message from buffer bytes to message structure
func ParseMsg(r io.Reader) (*Message, error) {
	bufferLength := make([]byte, 4)

	_, err := io.ReadFull(r, bufferLength)
	if err != nil {
		return nil, err
	}

	msgLength := binary.BigEndian.Uint32(bufferLength)

	// keep-connection alive
	if msgLength == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, msgLength)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      uint8(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	return &m, nil
}

func HaveMsg(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{
		ID:      Have,
		Payload: payload,
	}
}

func RequestMsg(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: Request, Payload: payload}
}

func ParseHaveMsg(msg *Message) (int, error) {

	// TODO: check whether it's a have msg
	if msg.ID != Have {
		log.Fatalf("Message Mismatch. Have and %d", msg.ID)
	}

	// TODO: payload validation

	completedIndex := int(binary.BigEndian.Uint32(msg.Payload))
	return completedIndex, nil
}

func ParsePieceMsg(msg *Message, index int, pieceBuffer []byte) (int, error) {
	if msg.ID != Bitfield {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", Bitfield, msg.ID)
	}
	// TODO: payload validation

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(pieceBuffer) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(pieceBuffer))
	}
	piece := msg.Payload[8:]
	if begin+len(piece) > len(pieceBuffer) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(piece), begin, len(pieceBuffer))
	}
	copy(pieceBuffer[begin:], piece)
	return len(piece), nil
}
