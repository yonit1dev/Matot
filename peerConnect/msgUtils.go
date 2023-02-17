package peerconnect

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

func (m *Message) BufferMsg() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	if m.Payload == nil {
		buffer := bytes.NewBuffer(make([]byte, 0, 5))
		binary.Write(buffer, binary.BigEndian, uint32(1))
		binary.Write(buffer, binary.BigEndian, m.ID)

		return buffer.Bytes()
	} else {
		msgLengthPrefix := uint32(len(m.Payload) + 1)

		buffer := bytes.NewBuffer(make([]byte, 0, msgLengthPrefix+4))

		binary.Write(buffer, binary.BigEndian, msgLengthPrefix)
		binary.Write(buffer, binary.BigEndian, m.ID)
		binary.Write(buffer, binary.BigEndian, m.Payload)

		return buffer.Bytes()
	}
}

func ReadMsg(r io.Reader) (*Message, error) {
	lengthPrefix := make([]byte, 4)

	_, err := io.ReadFull(r, lengthPrefix)
	if err != nil {
		return nil, err
	}

	msgLength := binary.BigEndian.Uint32(lengthPrefix)

	// Wait for connection - keep alive
	if msgLength == 0 {
		return nil, nil
	}

	// Msg buffer
	buffer := make([]byte, msgLength)
	_, err = io.ReadFull(r, buffer)
	if err != nil {
		return nil, err
	}

	if msgLength == 1 {
		msg := Message{
			ID: uint8(buffer[0]),
		}

		return &msg, nil
	} else if msgLength > 1 {
		msg := Message{
			ID:      uint8(buffer[0]),
			Payload: buffer[1:],
		}

		return &msg, nil
	} else {
		return nil, errors.New("no message recieved")
	}

}

func RequestMsgPayload(sp *SpecialMsg) *Message {
	reqPayload := make([]byte, 12) // 12 for the request structure
	binary.BigEndian.PutUint32(reqPayload[0:4], uint32(sp.Index))
	binary.BigEndian.PutUint32(reqPayload[4:8], uint32(sp.Begin))
	binary.BigEndian.PutUint32(reqPayload[8:12], uint32(sp.Length))

	return &Message{ID: Request, Payload: reqPayload}
}

func HaveMsgPayload(index uint32) *Message {
	havePayload := make([]byte, 4) // index of piece
	binary.BigEndian.PutUint32(havePayload, index)

	return &Message{ID: Have, Payload: havePayload}
}

func RecieveHaveMsg(m *Message) (int, error) {
	if m.ID != Have {
		log.Printf("Wrong message ID: %d", m.ID)
		return 0, errors.New("wrong msg")
	}

	if len(m.Payload) != 4 {
		log.Printf("Wrong payload length: %d", len(m.Payload))
		return 0, errors.New("wrong msg")
	}

	pieceIndex := int(binary.BigEndian.Uint32(m.Payload))

	return pieceIndex, nil
}

func RecievePieceMsg(reqIndex int, pieceBuf []byte, m *Message) (int, error) {
	if m.ID != Piece {
		log.Printf("Wrong message ID. Expected: %d, got: %d", Piece, m.ID)
		return 0, errors.New("wrong msg")
	}

	pieceIndex := int(binary.BigEndian.Uint32(m.Payload))

	if pieceIndex != reqIndex {
		log.Printf("Wrong piece index.  Expected: %d, got: %d", reqIndex, pieceIndex)
	}

	begin := int(binary.BigEndian.Uint32(m.Payload[4:8]))
	if begin >= len(pieceBuf) {
		log.Printf("Begin offset too high. %d >= %d", begin, len(pieceBuf))
		return 0, errors.New("wrong msg")
	}
	data := m.Payload[8:]
	if begin+len(data) > len(pieceBuf) {
		log.Printf("Data too long [%d] for offset %d with length %d", len(data), begin, len(pieceBuf))
		return 0, errors.New("wrong msg")
	}
	copy(pieceBuf[begin:], data)

	return len(data), nil
}

func RecieveBitfieldMsg(conn net.Conn) (BitFieldType, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	bitfieldMsg, err := ReadMsg(conn)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	if bitfieldMsg == nil {
		return nil, errors.New("wrong message - bitfield")
	}

	if bitfieldMsg.ID != BitField {
		log.Print("Wrong message ID for bitfield.")
		return nil, errors.New("wrong message - bitfield")
	}

	return bitfieldMsg.Payload, nil
}
